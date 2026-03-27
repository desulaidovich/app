package main

import (
	"context"
	"flag"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/desulaidovich/app/config"
	"github.com/desulaidovich/app/internal/migrator"
	"github.com/desulaidovich/app/internal/postgres"
	"github.com/desulaidovich/app/pkg/env"
	"github.com/desulaidovich/app/pkg/log"
)

func main() {
	ctx := context.Background()

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

	cmd := flag.String("cmd", "", "Команда на выполнение")
	param := flag.String("param", "", "Параметры (up/down/status/create)")
	name := flag.String("name", "", "Название миграции")

	flag.Parse()

	if *cmd == "" {
		panic("Команда не указана")
	}

	db, err := postgres.New(ctx,
		postgres.WithDSN(cfg.DSN()),
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

	m, err := migrator.New(cfg.DSN())
	if err != nil {
		panic("failed to init migrator: " + err.Error())
	}

	commands := map[string]func() error{
		"migrate": func() error {
			return handleMigrate(ctx, m, logger, *param, *name)
		},
	}

	handler, exists := commands[*cmd]
	if !exists {
		panic("unknown command: " + *cmd)
	}

	if err := handler(); err != nil {
		panic("failed to run command: " + err.Error())
	}
}

func handleMigrate(ctx context.Context, m *migrator.Migrator, logger log.Logger, param, name string) error {
	if param == "" {
		return fmt.Errorf("empty migration parameter")
	}

	migrateCommands := map[string]func() error{
		"create": func() error {
			if name == "" {
				return fmt.Errorf("empty migration name")
			}

			if err := m.Create(name); err != nil {
				return fmt.Errorf("failed to create migration: %w", err)
			}

			logger.With(map[string]any{
				"name": name,
			}).Info("Migration created successfully")
			return nil
		},
		"up": func() error {
			if err := m.Up(ctx); err != nil {
				return fmt.Errorf("failed to apply migrations: %w", err)
			}

			logger.Info("Migrations applied successfully")
			return nil
		},
		"down": func() error {
			if err := m.Down(ctx); err != nil {
				return fmt.Errorf("failed to rollback migrations: %w", err)
			}

			logger.Info("Migrations rolled back successfully")
			return nil
		},
		"status": func() error {
			if err := m.Status(ctx); err != nil {
				return fmt.Errorf("failed to get migration status: %w", err)
			}
			return nil
		},
		"version": func() error {
			version, err := m.Version(ctx)
			if err != nil {
				return fmt.Errorf("failed to get migration version: %w", err)
			}

			logger.With(map[string]any{
				"version": version,
			}).Info("Current migration version")
			return nil
		},
	}

	handler, exists := migrateCommands[param]
	if !exists {
		return fmt.Errorf("unknown migration command: %s", param)
	}

	return handler()
}
