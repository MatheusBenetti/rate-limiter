package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/MatheusBenetti/rate-limiter/internal/dto"
	"github.com/MatheusBenetti/rate-limiter/internal/entity"
	"github.com/redis/go-redis/v9"
)

type APIKeyRedis struct {
	redisCli *redis.Client
}

func NewAPIKeyRedis(redisCli *redis.Client) *APIKeyRedis {
	return &APIKeyRedis{redisCli: redisCli}
}

func (at *APIKeyRedis) Save(ctx context.Context, key *entity.ApiKey) (string, error) {
	req := dto.Input{
		MaxReq:        key.RateLimiter.MaxReq,
		TimeWindow:    key.RateLimiter.TimeWindow,
		BlockDuration: key.BlockDuration,
	}

	jsonReq, marErr := json.Marshal(req)
	if marErr != nil {
		log.Println("error marshaling API Key")
		return "", marErr
	}

	if redisErr := at.redisCli.Set(
		ctx,
		key.Value(),
		jsonReq,
		0,
	).Err(); redisErr != nil {
		log.Println("error inserting API Key value")
		return "", redisErr
	}

	return key.Value(), nil
}

func (at *APIKeyRedis) Get(ctx context.Context, value string) (*entity.ApiKey, error) {
	val, getErr := at.redisCli.Get(ctx, value).Result()
	if getErr != nil {
		return &entity.ApiKey{}, getErr
	}

	var apiKeyConfigDB dto.Input
	if err := json.Unmarshal([]byte(val), &apiKeyConfigDB); err != nil {
		log.Println("API key configuration marshall error")
		return &entity.ApiKey{}, err
	}

	return &entity.ApiKey{
		BlockDuration: apiKeyConfigDB.BlockDuration,
		RateLimiter: entity.RateLimiter{
			TimeWindow: apiKeyConfigDB.TimeWindow,
			MaxReq:     apiKeyConfigDB.MaxReq,
		},
	}, nil
}

func (at *APIKeyRedis) UpsertRequest(ctx context.Context, key string, rl *entity.RateLimiter) error {
	req := dto.ApiKeyReqDb{
		MaxReq:     rl.MaxReq,
		TimeWindow: rl.TimeWindow,
		Req: func() []int64 {
			reqInt := make([]int64, 0)
			for _, r := range rl.Req {
				reqInt = append(reqInt, r.Unix())
			}
			return reqInt
		}(),
	}

	jsonReq, marErr := json.Marshal(req)
	if marErr != nil {
		log.Println("error marshaling API Key")
		return marErr
	}

	redisErr := at.redisCli.Set(ctx, createAPIKeyRatePrefix(key), jsonReq, 0).Err()
	if redisErr != nil {
		log.Println("error inserting API Key value")
		return redisErr
	}

	return nil
}

func (at *APIKeyRedis) SaveBlockedDuration(ctx context.Context, key string, BlockedDuration int64) error {
	if redisErr := at.redisCli.Set(
		ctx,
		createAPIKeyDurationPrefix(key),
		entity.StatusApiKeyBlock,
		time.Second*time.Duration(BlockedDuration),
	).Err(); redisErr != nil {
		log.Println("error inserting SaveBlockedDuration on API Key")
		return redisErr
	}

	return nil
}

func (at *APIKeyRedis) GetBlockedDuration(ctx context.Context, key string) (string, error) {
	val, getErr := at.redisCli.Get(ctx, createAPIKeyDurationPrefix(key)).Result()
	if errors.Is(getErr, redis.Nil) {
		log.Println("API key does not exist")
		return "", nil
	}
	if getErr != nil {
		return "", getErr
	}

	return val, nil
}

// GetRequest reads the stored array of request
func (at *APIKeyRedis) GetRequest(ctx context.Context, key string) (*entity.RateLimiter, error) {
	val, getErr := at.redisCli.Get(ctx, createAPIKeyRatePrefix(key)).Result()
	if errors.Is(getErr, redis.Nil) {
		log.Println("INFO: GetRequest API key does not exist")
		return &entity.RateLimiter{
			Req:        make([]time.Time, 0),
			TimeWindow: 0,
			MaxReq:     0,
		}, nil
	}
	if getErr != nil {
		return nil, getErr
	}

	var rateLimiter dto.ApiKeyReqDb
	if err := json.Unmarshal([]byte(val), &rateLimiter); err != nil {
		log.Println("API key RateLimiter unmarshal error")
		return &entity.RateLimiter{}, err
	}

	return &entity.RateLimiter{
		Req: func() []time.Time {
			reqTimeStamp := make([]time.Time, 0)
			for _, rr := range rateLimiter.Req {
				reqTimeStamp = append(reqTimeStamp, time.Unix(rr, 0))
			}
			return reqTimeStamp
		}(),
		TimeWindow: rateLimiter.TimeWindow,
		MaxReq:     rateLimiter.MaxReq,
	}, nil
}

func createAPIKeyDurationPrefix(key string) string {
	return fmt.Sprintf("%s_%s", entity.ApiKeyBlockDuration, key)
}

func createAPIKeyRatePrefix(key string) string {
	return fmt.Sprintf("%s_%s", entity.ApiKeyRateKey, key)
}
