package app

import (
	"context"
	"errors"

	"github.com/desulaidovich/app/config"
	"github.com/desulaidovich/app/pkg/log"
)

type App struct {
	cfg     *config.Config
	log     log.Logger
	version string
	build   string
}

type Option func(*App) error

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

	return app, nil
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
