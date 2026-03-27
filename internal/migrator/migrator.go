package migrator

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

const (
	Type      = "go"
	Create    = "up"
	Delete    = "down"
	Directory = "./migrations"
)

type Migrator struct {
	db *sql.DB
}

func New(dsn string) (*Migrator, error) {
	if err := goose.SetDialect("postgres"); err != nil {
		return nil, fmt.Errorf("migrator: failed to set dialect: %w", err)
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("migrator: failed to open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("migrator: failed to ping database: %w", err)
	}

	return &Migrator{
		db: db,
	}, nil
}

func (m *Migrator) Create(name string) error {
	if err := goose.Create(m.db, Directory, name, Type); err != nil {
		return fmt.Errorf("migrator: failed to create migration: %w", err)
	}

	return nil
}

func (m *Migrator) Up(ctx context.Context) error {
	if err := goose.RunContext(ctx, Create, m.db, Directory); err != nil {
		return fmt.Errorf("migrator: failed to up: %w", err)
	}
	return nil
}

func (m *Migrator) Down(ctx context.Context) error {
	if err := goose.RunContext(ctx, Delete, m.db, Directory); err != nil {
		return fmt.Errorf("migrator: failed to down: %w", err)
	}
	return nil
}

func (m *Migrator) Status(ctx context.Context) error {
	if err := goose.RunContext(ctx, "status", m.db, Directory); err != nil {
		return fmt.Errorf("migrator: failed to get status: %w", err)
	}
	return nil
}

func (m *Migrator) Version(_ context.Context) (int64, error) {
	return goose.GetDBVersion(m.db)
}

func (m *Migrator) Close() error {
	return m.db.Close()
}
