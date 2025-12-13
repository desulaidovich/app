package main

import (
	"context"

	"github.com/desulaidovich/config"
	"github.com/desulaidovich/internal/app"
	"github.com/desulaidovich/pkg/env"
	"github.com/desulaidovich/pkg/log"
	"github.com/desulaidovich/pkg/runner"
)

func main() {
	cfg, err := env.New(config.Config{}).FromENV().Parse()
	if err != nil {
		panic("parse env config: " + err.Error())
	}

	log := log.New(cfg.LogLevel)

	app := app.New(&cfg, log)

	if err := runner.New(app).Run(context.Background()); err != nil {
		panic("run app: " + err.Error())
	}
}
