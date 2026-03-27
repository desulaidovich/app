package postgres

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultMaxConns          = 25
	defaultMinConns          = 5
	defaultMaxConnLifetime   = 5 * time.Minute
	defaultMaxConnIdleTime   = 5 * time.Minute
	defaultHealthCheckPeriod = 1 * time.Minute
	defaultConnectTimeout    = 5 * time.Second
)

type Pool struct {
	*pgxpool.Pool
	config *Config
}

type Config struct {
	DSN               string
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
	ConnectTimeout    time.Duration
}

type Option func(*Config) error

func WithMaxConns(maxConns int32) Option {
	return func(c *Config) error {
		if maxConns <= 0 {
			return errors.New("maxConns must be positive")
		}
		c.MaxConns = maxConns
		return nil
	}
}

func WithMinConns(minConns int32) Option {
	return func(c *Config) error {
		if minConns < 0 {
			return errors.New("minConns cannot be negative")
		}
		c.MinConns = minConns
		return nil
	}
}

func WithMaxConnLifetime(lifetime time.Duration) Option {
	return func(c *Config) error {
		if lifetime <= 0 {
			return errors.New("maxConnLifetime must be positive")
		}
		c.MaxConnLifetime = lifetime
		return nil
	}
}

func WithMaxConnIdleTime(idleTime time.Duration) Option {
	return func(c *Config) error {
		if idleTime <= 0 {
			return errors.New("maxConnIdleTime must be positive")
		}
		c.MaxConnIdleTime = idleTime
		return nil
	}
}

func WithHealthCheckPeriod(period time.Duration) Option {
	return func(c *Config) error {
		if period <= 0 {
			return errors.New("healthCheckPeriod must be positive")
		}
		c.HealthCheckPeriod = period
		return nil
	}
}

func WithConnectTimeout(timeout time.Duration) Option {
	return func(c *Config) error {
		if timeout <= 0 {
			return errors.New("connectTimeout must be positive")
		}
		c.ConnectTimeout = timeout
		return nil
	}
}

func WithDSN(dsn string) Option {
	return func(c *Config) error {
		if dsn == "" {
			return errors.New("dsn cannot be empty")
		}
		c.DSN = dsn
		return nil
	}
}

func WithSSLMode(mode string) Option {
	return func(c *Config) error {
		if !slices.Contains([]string{"disable", "require", "verify-ca", "verify-full"}, mode) {
			return fmt.Errorf("unsupported SSL mode: %s", mode)
		}
		c.DSN += "&sslmode=" + mode
		return nil
	}
}

func New(ctx context.Context, opts ...Option) (*Pool, error) {
	cfg := &Config{
		MaxConns:          defaultMaxConns,
		MinConns:          defaultMinConns,
		MaxConnLifetime:   defaultMaxConnLifetime,
		MaxConnIdleTime:   defaultMaxConnIdleTime,
		HealthCheckPeriod: defaultHealthCheckPeriod,
		ConnectTimeout:    defaultConnectTimeout,
	}

	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, fmt.Errorf("postgres option: %w", err)
		}
	}

	if cfg.DSN == "" {
		return nil, errors.New("DSN is required: use WithDSN option")
	}

	pgxCfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	pgxCfg.MaxConns = cfg.MaxConns
	pgxCfg.MinConns = cfg.MinConns
	pgxCfg.MaxConnLifetime = cfg.MaxConnLifetime
	pgxCfg.MaxConnIdleTime = cfg.MaxConnIdleTime
	pgxCfg.HealthCheckPeriod = cfg.HealthCheckPeriod
	pgxCfg.ConnConfig.ConnectTimeout = cfg.ConnectTimeout

	pool, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Pool{Pool: pool, config: cfg}, nil
}

func (p *Pool) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}
