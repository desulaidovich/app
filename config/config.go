package config

import "fmt"

type Config struct {
	App struct {
		Name  string `env:"NAME"`
		Env   string `env:"ENV"`
		Debug bool   `env:"DEBUG"`
	} `env:"APP"`

	Database struct {
		Host string `env:"HOST"`
		Port int    `env:"PORT"`
		Name string `env:"NAME"`
		User struct {
			Name     string `env:"NAME"`
			Password string `env:"PASSWORD"`
		}
	} `env:"DATABASE"`

	LogLevel string `env:"LOG_LEVEL"`
}

func (cfg Config) ConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?application_name=%s",
		cfg.Database.User.Name,
		cfg.Database.User.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.App.Name,
	)
}
