package main

import (
	"context"
	"syscall"
	"time"

	"github.com/desulaidovich/app/config"
	"github.com/desulaidovich/app/internal/app"
	"github.com/desulaidovich/app/internal/postgres"
	"github.com/desulaidovich/app/pkg/env"
	"github.com/desulaidovich/app/pkg/log"
	"github.com/desulaidovich/app/pkg/runner"
)

// Build-time variables, injected via -ldflags.
var (
	version = "dev"
	build   = "unknown"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
		panic("failed to create logger: " + err.Error())
	}

	db, err := postgres.New(ctx,
		postgres.WithDSN(cfg.DSN()),
		postgres.WithMaxConns(cfg.Database.Pool.MaxConns),
		postgres.WithMinConns(cfg.Database.Pool.MinConns),
		postgres.WithMaxConnLifetime(cfg.Database.Pool.MaxConnLifetime),
		postgres.WithMaxConnIdleTime(cfg.Database.Pool.MaxConnIdleTime),
		postgres.WithConnectTimeout(cfg.Database.Pool.ConnectTimeout),
		postgres.WithHealthCheckPeriod(2*time.Minute),
		postgres.WithSSLMode(cfg.Database.SSLMode),
	)
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}
	defer db.Close()

	application, err := app.New(
		app.WithAppName(cfg.App.Name),
		app.WithVersion(version, build),
		app.WithConfig(&cfg),
		app.WithLogger(logger),
		app.WithPostgres(db),
	)
	if err != nil {
		panic("failed to run migrations: " + err.Error())
	}

	r, err := runner.New(
		runner.WithHandler(application),
		runner.WithSignals(syscall.SIGINT, syscall.SIGTERM),
		runner.WithStopTimeout(5*time.Second),
	)
	if err != nil {
		panic("failed to create runner: " + err.Error())
	}

	if err := r.Run(ctx); err != nil {
		panic("runner exited with error: " + err.Error())
	}
}
