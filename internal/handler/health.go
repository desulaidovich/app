package handler

import (
	"context"

	healthv1 "github.com/desulaidovich/app/api/health/v1"
)

type HealthHandler struct {
	healthv1.UnimplementedHealthServiceServer
	version string
	build   string
}

func NewHealthHandler(version, build string) *HealthHandler {
	return &HealthHandler{
		version: version,
		build:   build,
	}
}

func (h *HealthHandler) Health(_ context.Context, _ *healthv1.HealthRequest) (*healthv1.HealthResponse, error) {
	return &healthv1.HealthResponse{
		Status:  "ok",
		Version: h.version,
		Build:   h.build,
	}, nil
}
