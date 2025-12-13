package app

import (
	"context"

	"github.com/desulaidovich/config"
	"github.com/desulaidovich/pkg/log"
)

type App struct {
	cfg *config.Config
	log *log.Logger
}

func New(cfg *config.Config, log *log.Logger) *App {
	return &App{
		cfg: cfg,
		log: log,
	}
}

func (app *App) Start(ctx context.Context) error {
	status := map[bool]string{
		true:  "enabled",
		false: "disabled",
	}

	app.log.
		With("name", app.cfg.App.Name).
		With("mode", app.cfg.App.Env).
		With("debug", status[app.cfg.App.Debug]).
		Info("Application started")

	return nil
}

func (app *App) Stop(ctx context.Context) error {
	app.log.Info("Application stopped")

	return nil
}
