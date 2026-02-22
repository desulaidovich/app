package postgres

import (
	"context"
	"errors"
	"fmt"
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

func WithSSLMode(mode string) Option {
	return func(c *Config) error {
		validModes := map[string]bool{
			"disable":     true,
			"require":     true,
			"verify-ca":   true,
			"verify-full": true,
		}
		if !validModes[mode] {
			return fmt.Errorf("unsupported SSL mode: %s", mode)
		}

		c.DSN = addOrUpdateParam(c.DSN, "sslmode", mode)
		return nil
	}
}

func New(ctx context.Context, dsn string, opts ...Option) (*Pool, error) {
	if dsn == "" {
		return nil, errors.New("dsn cannot be empty")
	}

	config := &Config{
		DSN:               dsn,
		MaxConns:          defaultMaxConns,
		MinConns:          defaultMinConns,
		MaxConnLifetime:   defaultMaxConnLifetime,
		MaxConnIdleTime:   defaultMaxConnIdleTime,
		HealthCheckPeriod: defaultHealthCheckPeriod,
		ConnectTimeout:    defaultConnectTimeout,
	}

	for _, opt := range opts {
		if err := opt(config); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	pgxConfig, err := pgxpool.ParseConfig(config.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	pgxConfig.MaxConns = config.MaxConns
	pgxConfig.MinConns = config.MinConns
	pgxConfig.MaxConnLifetime = config.MaxConnLifetime
	pgxConfig.MaxConnIdleTime = config.MaxConnIdleTime
	pgxConfig.HealthCheckPeriod = config.HealthCheckPeriod
	pgxConfig.ConnConfig.ConnectTimeout = config.ConnectTimeout

	pool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Pool{
		Pool:   pool,
		config: config,
	}, nil
}

func (p *Pool) Config() Config {
	return *p.config
}

func (p *Pool) Stats() *pgxpool.Stat {
	return p.Stat()
}

func (p *Pool) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}

func addOrUpdateParam(dsn, key, value string) string {
	param := fmt.Sprintf("%s=%s", key, value)

	if dsn == "" {
		return param
	}

	return dsn + "&" + param
}
