package config

import (
	"fmt"
)

type Config struct {
	App struct {
		Name  string `env:"NAME"`
		Env   string `env:"ENV,default=development"`
		Debug bool   `env:"DEBUG"`
	} `env:"APP"`

	HTTP struct {
		Port string `env:"PORT,default=8080"`
	} `env:"HTTP"`

	Database struct {
		Host string `env:"HOST"`
		Port int    `env:"PORT,default=5432"`
		Name string `env:"NAME"`
		User struct {
			Name     string `env:"NAME"`
			Password string `env:"PASSWORD"`
		}
	} `env:"DATABASE"`

	Log struct {
		Level      string `env:"LEVEL,default=debug"`
		Format     string `env:"FORMAT,default=text"`
		TimeFormat string `env:"TIMEFORMAT,default=2006-01-02T15:04:05Z07:00"`
	} `env:"LOG"`
}

func (cfg Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?application_name=%s",
		cfg.Database.User.Name,
		cfg.Database.User.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.App.Name,
	)
}
