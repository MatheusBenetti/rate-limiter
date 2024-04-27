package middleware

import (
	"net/http"

	"github.com/MatheusBenetti/rate-limiter/config"
	"github.com/MatheusBenetti/rate-limiter/internal/entity"
	"github.com/redis/go-redis/v9"
)

type Middleware struct {
	RedisClient *redis.Client
	Config      *config.Config
}

func (m *Middleware) RateLimiter(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get(entity.ApiKeyHeader)
			strategy := Factory(apiKey, m)
			if err := strategy.Execute(w, r); err != nil {
				return
			}

			next.ServeHTTP(w, r)
		},
	)
}
