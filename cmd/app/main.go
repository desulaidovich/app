package main

import (
	"context"

	"github.com/desulaidovich/app/config"
	"github.com/desulaidovich/app/internal/app"
	"github.com/desulaidovich/app/internal/postgres"
	"github.com/desulaidovich/app/pkg/env"
	"github.com/desulaidovich/app/pkg/log"
	"github.com/desulaidovich/app/pkg/runner"
)

func main() {
	ctx := context.Background()

	var cfg config.Config
	if err := env.LoadFile(&cfg, ".env"); err != nil {
		panic("failed to load config: " + err.Error())
	}

	db, err := postgres.New(ctx, cfg.DSN(),
		postgres.WithMaxConns(cfg.Database.Pool.MaxConns),
		postgres.WithMinConns(cfg.Database.Pool.MinConns),
		postgres.WithMaxConnLifetime(cfg.Database.Pool.MaxConnLifetime),
		postgres.WithMaxConnIdleTime(cfg.Database.Pool.MaxConnIdleTime),
		postgres.WithConnectTimeout(cfg.Database.Pool.ConnectTimeout),
		postgres.WithSSLMode(cfg.Database.SSLMode),
	)
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}
	defer db.Close()

	logger, err := log.New(
		log.WithLevel(cfg.Log.Level),
		log.WithFormat(cfg.Log.Format),
		log.WithTimeFormat(cfg.Log.TimeFormat),
	)
	if err != nil {
		panic("failed to create logger" + err.Error())
	}

	logger.With(map[string]any{
		"config_json": cfg,
	}).Debug("Loaded conrfig")

	app := app.New(
		app.WithVersion("1.0.0", "123qwe"),
		app.WithConfig(&cfg),
		app.WithLogger(logger),
	)

	if err := runner.New(app).Run(context.Background()); err != nil {
		panic("failed to run application: " + err.Error())
	}
}
