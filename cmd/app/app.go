package main

import (
	"context"

	"github.com/Desulaidovich/app/config"
	"github.com/Desulaidovich/app/internal/app"
	"github.com/Desulaidovich/app/pkg/env"
	"github.com/Desulaidovich/app/pkg/log"
	"github.com/Desulaidovich/app/pkg/runner"
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
