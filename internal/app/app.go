package app

import (
	"context"

	"github.com/desulaidovich/app/config"
	"github.com/desulaidovich/app/pkg/log"
)

type App struct {
	cfg     *config.Config
	log     log.Logger
	version string
	build   string
}

type Option func(*App)

func WithVersion(version, build string) Option {
	return func(a *App) {
		a.version = version
		a.build = build
	}
}

func WithConfig(cfg *config.Config) Option {
	return func(a *App) {
		a.cfg = cfg
	}
}

func WithLogger(logger log.Logger) Option {
	return func(a *App) {
		a.log = logger
	}
}

func New(opts ...Option) *App {
	app := new(App)

	for _, opt := range opts {
		opt(app)
	}

	return app
}

func (app *App) Start(ctx context.Context) error {
	status := map[bool]string{
		true:  "enabled",
		false: "disabled",
	}

	app.log.With(map[string]any{
		"name":    app.cfg.App.Name,
		"mode":    app.cfg.App.Env,
		"debug":   status[app.cfg.App.Debug],
		"version": app.version,
		"build":   app.build,
	}).Info("Application started")

	return nil
}

func (app *App) Stop(ctx context.Context) error {
	app.log.Info("Application stopped")
	return nil
}
