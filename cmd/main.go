package main

import (
	"context"
	"net"
	"net/http"
	"uber-go/pkg/handlers"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewHTTPServer(lc fx.Lifecycle, mux *http.ServeMux, log *zap.Logger) *http.Server {
	srv := &http.Server{Addr: ":8080", Handler: mux}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			log.Info("Starting HTTP server at", zap.String("addr", srv.Addr))
			go srv.Serve(ln)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})
	return srv
}

func NewServeMux(routes []handlers.Route) *http.ServeMux {
	mux := http.NewServeMux()
	for _, route := range routes {
		mux.Handle(route.Pattern(), route)
	}
	return mux
}

// AsRoute annotates the given constructor to state that
// it provides a route to the "routes" group.
func AsRoute(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(handlers.Route)),
		fx.ResultTags(`group:"routes"`),
	)
}

func main() {
	fx.New(
		// fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
		// 	return &fxevent.ZapLogger{Logger: log}
		// }),
		fx.Provide(
			NewHTTPServer,
			fx.Annotate(
				NewServeMux,
				fx.ParamTags(`group:"routes"`),
			),
			AsRoute(handlers.NewEchoHandler),
			AsRoute(handlers.NewHelloHandler),
			zap.NewExample,
		),
		fx.Invoke(func(*http.Server) {}),
	).Run()
}
