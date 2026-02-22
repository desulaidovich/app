package main

import (
	"context"

	"github.com/desulaidovich/app/config"
	"github.com/desulaidovich/app/internal/app"
	"github.com/desulaidovich/app/pkg/env"
	"github.com/desulaidovich/app/pkg/log"
	"github.com/desulaidovich/app/pkg/runner"
)

func main() {
	var cfg config.Config
	if err := env.Load(&cfg, ".env"); err != nil {
		panic("failed to load config: " + err.Error())
	}

	logger, err := log.New(
		log.WithLevel(cfg.Log.Level),
		log.WithFormat(cfg.Log.Format),
		log.WithTimeFormat(cfg.Log.TimeFormat),
	)
	if err != nil {
		panic("failed to create logger" + err.Error())
	}

	app := app.New(
		app.WithVersion("1.0.0", "123qwe"),
		app.WithConfig(&cfg),
		app.WithLogger(logger),
	)

	if err := runner.New(app).Run(context.Background()); err != nil {
		panic("failed to run application: " + err.Error())
	}
}
