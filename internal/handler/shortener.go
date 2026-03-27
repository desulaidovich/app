package handler

import (
	"context"

	shortenerv1 "github.com/desulaidovich/app/api/shortener/v1"
)

type ShortenerHandler struct {
	shortenerv1.UnimplementedShortenerServiceServer
}

func NewShortenerHandler() *ShortenerHandler {
	return &ShortenerHandler{}
}

func (h *HealthHandler) Shorten(ctx context.Context, req *shortenerv1.ShortenRequest) (*shortenerv1.ShortenResponse, error) {
	return nil, nil
}

func (h *HealthHandler) GetOriginal(ctx context.Context, req *shortenerv1.GetOriginalRequest) (*shortenerv1.GetOriginalResponse, error) {
	return nil, nil
}

func (h *HealthHandler) Redirect(ctx context.Context, req *shortenerv1.RedirectRequest) (*shortenerv1.RedirectResponse, error) {
	return nil, nil
}

func (h *HealthHandler) GetStats(ctx context.Context, req *shortenerv1.GetStatsRequest) (*shortenerv1.GetStatsResponse, error) {
	return nil, nil
}

func (h *HealthHandler) Update(ctx context.Context, req *shortenerv1.UpdateRequest) (*shortenerv1.UpdateResponse, error) {
	return nil, nil
}

func (h *HealthHandler) Delete(ctx context.Context, req *shortenerv1.DeleteRequest) (*shortenerv1.DeleteResponse, error) {
	return nil, nil
}

func (h *HealthHandler) ListUserLinks(ctx context.Context, req *shortenerv1.ListUserLinksRequest) (*shortenerv1.ListUserLinksResponse, error) {
	return nil, nil
}
