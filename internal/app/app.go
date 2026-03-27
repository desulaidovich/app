package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	healthv1 "github.com/desulaidovich/app/api/health/v1"
	"github.com/desulaidovich/app/config"
	"github.com/desulaidovich/app/internal/grpcserver"
	"github.com/desulaidovich/app/internal/handler"
	"github.com/desulaidovich/app/internal/middleware"
	"github.com/desulaidovich/app/internal/postgres"
	"github.com/desulaidovich/app/pkg/log"
)

type App struct {
	cfg     *config.Config
	log     log.Logger
	db      *postgres.Pool
	grpcSrv *grpcserver.Server
	httpSrv *http.Server
	name    string
	version string
	build   string
}

type Option func(*App) error

func WithAppName(name string) Option {
	return func(a *App) error {
		if name == "" {
			return errors.New("app name cannot be empty")
		}
		a.name = name
		return nil
	}
}

func WithVersion(version, build string) Option {
	return func(a *App) error {
		if version == "" {
			return errors.New("version cannot be empty")
		}
		if build == "" {
			return errors.New("build cannot be empty")
		}
		a.version = version
		a.build = build
		return nil
	}
}

func WithConfig(cfg *config.Config) Option {
	return func(a *App) error {
		if cfg == nil {
			return errors.New("config cannot be nil")
		}
		a.cfg = cfg
		return nil
	}
}

func WithLogger(logger log.Logger) Option {
	return func(a *App) error {
		if logger == nil {
			return errors.New("logger cannot be nil")
		}
		a.log = logger
		return nil
	}
}

func WithPostgres(db *postgres.Pool) Option {
	return func(a *App) error {
		if db == nil {
			return errors.New("postgres pool cannot be nil")
		}
		a.db = db
		return nil
	}
}

func New(opts ...Option) (*App, error) {
	app := new(App)

	for _, opt := range opts {
		if err := opt(app); err != nil {
			return nil, err
		}
	}

	if app.cfg == nil {
		return nil, errors.New("config is required")
	}
	if app.log == nil {
		return nil, errors.New("logger is required")
	}
	if app.db == nil {
		return nil, errors.New("postgres is required")
	}

	app.grpcSrv = grpcserver.New(
		net.JoinHostPort("", app.cfg.GRPC.Port),
		grpc.ChainUnaryInterceptor(
			middleware.GRPCRecovery(app.log),
			middleware.GRPCLogging(app.log),
		),
	)

	if app.cfg.App.Debug {
		reflection.Register(app.grpcSrv.Server())
	}

	healthHandler := handler.NewHealthHandler(app.version, app.build)
	healthv1.RegisterHealthServiceServer(app.grpcSrv.Server(), healthHandler)

	gwMux := runtime.NewServeMux()
	if err := healthv1.RegisterHealthServiceHandlerServer(context.Background(), gwMux, healthHandler); err != nil {
		return nil, fmt.Errorf("failed to register health service handler: %w", err)
	}

	app.httpSrv = &http.Server{
		Addr:              net.JoinHostPort("", app.cfg.HTTP.Port),
		Handler:           middleware.Chain(gwMux, middleware.Logging(app.log), middleware.CORS),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return app, nil
}

func (app *App) Start(ctx context.Context) error {
	app.log.With(map[string]any{
		"name":      app.cfg.App.Name,
		"mode":      app.cfg.App.Env,
		"version":   app.version,
		"build":     app.build,
		"grpc_addr": app.grpcSrv.Addr(),
		"http_addr": app.httpSrv.Addr,
	}).Info("Application started")

	errCh := make(chan error, 1)

	go func() {
		if err := app.grpcSrv.Start(ctx); err != nil {
			errCh <- fmt.Errorf("grpc server start error: %w", err)
		}
	}()

	go func() {
		if err := app.httpSrv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("http server start error: %w", err)
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return nil
	}
}

func (app *App) Stop(ctx context.Context) error {
	app.log.Info("Application stopping")

	if err := app.grpcSrv.Stop(ctx); err != nil {
		return fmt.Errorf("could not stop grpc server: %w", err)
	}

	if err := app.httpSrv.Shutdown(ctx); err != nil {
		return fmt.Errorf("could not stop http server: %w", err)
	}

	return nil
}
