package entity

import (
	"context"
)

//go:generate mockgen -source repository.go -destination mock/repository_mock.go -package mock
type commonRepository interface {
	UpsertRequest(ctx context.Context, key string, rl *RateLimiter) error

	SaveBlockedDuration(ctx context.Context, key string, BlockedDuration int64) error

	GetBlockedDuration(ctx context.Context, key string) (string, error)

	GetRequest(ctx context.Context, key string) (*RateLimiter, error)
}

type ApiKeyRepository interface {
	Save(ctx context.Context, key *ApiKey) (string, error)

	Get(ctx context.Context, value string) (*ApiKey, error)

	commonRepository
}

type IPRepository interface {
	commonRepository
}
