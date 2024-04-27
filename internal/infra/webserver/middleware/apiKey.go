package middleware

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/MatheusBenetti/rate-limiter/internal/dto"
	"github.com/MatheusBenetti/rate-limiter/internal/entity"
	"github.com/MatheusBenetti/rate-limiter/internal/infra/database"
	"github.com/MatheusBenetti/rate-limiter/internal/usecase"
	"github.com/redis/go-redis/v9"
)

type APIKeyMiddleware struct {
	RedisClient *redis.Client
	ApiKey      string
}

func (tk *APIKeyMiddleware) Execute(w http.ResponseWriter, r *http.Request) error {
	tkDB := database.NewAPIKeyRedis(tk.RedisClient)
	tkReq := usecase.NewRegisterAPIKeyUseCase(tkDB)
	execute, execErr := tkReq.Execute(r.Context(), dto.ApiKeyReq{
		Value:     tk.ApiKey,
		TimeAdded: time.Now(),
	})
	if errors.Is(execErr, entity.ErrApiKeyAmountReq) {
		log.Printf("Error executing ErrRateLimiterMaxRequests: %s\n", execErr.Error())
		http.Error(w, execErr.Error(), http.StatusTooManyRequests)
		return execErr
	}
	if execErr != nil {
		log.Printf("Error executing NewRegisterAPIKeyUseCase: %s\n", execErr.Error())
		http.Error(w, execErr.Error(), http.StatusInternalServerError)
		return execErr
	}

	if !execute.Allow {
		log.Printf("Too many request: %s\n", entity.ErrApiKeyAmountReq.Error())
		http.Error(w, entity.ErrApiKeyAmountReq.Error(), http.StatusTooManyRequests)
		return errors.New("too many request")
	}

	return nil
}
