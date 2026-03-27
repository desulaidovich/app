package handler

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	shortenerv1 "github.com/desulaidovich/app/api/shortener/v1"
)

type ShortenerHandler struct {
	shortenerv1.UnimplementedShortenerServiceServer
}

func NewShortenerHandler() *ShortenerHandler {
	return &ShortenerHandler{}
}

func (h *ShortenerHandler) Shorten(ctx context.Context, req *shortenerv1.ShortenRequest) (*shortenerv1.ShortenResponse, error) {
	return &shortenerv1.ShortenResponse{
		ShortCode: "",
		ShortUrl:  "",
		IsCustom:  true,
		CreatedAt: timestamppb.New(time.Now()),
		ExpiresAt: timestamppb.New(time.Now().Add(24 * time.Hour)),
	}, nil
}

func (h *ShortenerHandler) GetOriginal(ctx context.Context, req *shortenerv1.GetOriginalRequest) (*shortenerv1.GetOriginalResponse, error) {
	return nil, nil
}

func (h *ShortenerHandler) Redirect(ctx context.Context, req *shortenerv1.RedirectRequest) (*shortenerv1.RedirectResponse, error) {
	return nil, nil
}

func (h *ShortenerHandler) GetStats(ctx context.Context, req *shortenerv1.GetStatsRequest) (*shortenerv1.GetStatsResponse, error) {
	return nil, nil
}

func (h *ShortenerHandler) Update(ctx context.Context, req *shortenerv1.UpdateRequest) (*shortenerv1.UpdateResponse, error) {
	return nil, nil
}

func (h *ShortenerHandler) Delete(ctx context.Context, req *shortenerv1.DeleteRequest) (*shortenerv1.DeleteResponse, error) {
	return nil, nil
}

func (h *ShortenerHandler) ListUserLinks(ctx context.Context, req *shortenerv1.ListUserLinksRequest) (*shortenerv1.ListUserLinksResponse, error) {
	return nil, nil
}
