package usecase

import (
	"context"
	"log"

	"github.com/MatheusBenetti/rate-limiter/internal/dto"
	"github.com/MatheusBenetti/rate-limiter/internal/entity"
)

type RegisterApiKey struct {
	apiRepository entity.ApiKeyRepository
}

func NewRegisterAPIKeyUseCase(
	apiRepository entity.ApiKeyRepository,
) *RegisterApiKey {
	return &RegisterApiKey{
		apiRepository: apiRepository,
	}
}

func (apk *RegisterApiKey) Execute(
	ctx context.Context,
	input dto.ApiKeyReq,
) (dto.ApiKeyAllow, error) {
	status, blockedErr := apk.apiRepository.GetBlockedDuration(ctx, input.Value)
	if blockedErr != nil {
		return dto.ApiKeyAllow{}, blockedErr
	}

	if status == entity.StatusApiKeyBlock {
		log.Println("API key is blocked due to exceeding the maximum number of requests")
		return dto.ApiKeyAllow{}, entity.ErrApiKeyAmountReq
	}

	apiKeyConfig, getErr := apk.apiRepository.Get(ctx, input.Value)
	if getErr != nil {
		log.Println("API key get error:", getErr.Error())
		return dto.ApiKeyAllow{}, getErr
	}

	rateLimReq, getReqErr := apk.apiRepository.GetRequest(ctx, input.Value)
	if getReqErr != nil {
		log.Printf("Error getting IP requests: %s \n", getReqErr.Error())
		return dto.ApiKeyAllow{}, getReqErr
	}

	rateLimReq.TimeWindow = apiKeyConfig.RateLimiter.TimeWindow
	rateLimReq.MaxReq = apiKeyConfig.RateLimiter.MaxReq
	if valErr := rateLimReq.Validate(); valErr != nil {
		log.Printf("Error validation in rate limiter: %s \n", valErr.Error())
		return dto.ApiKeyAllow{}, valErr
	}

	rateLimReq.AddReq(input.TimeAdded)
	isAllowed := rateLimReq.Allow(input.TimeAdded)
	if upsertErr := apk.apiRepository.UpsertRequest(ctx, input.Value, rateLimReq); upsertErr != nil {
		log.Printf("Error updating/inserting rate limit: %s \n", upsertErr.Error())
		return dto.ApiKeyAllow{}, upsertErr
	}

	if !isAllowed {
		if saveErr := apk.apiRepository.SaveBlockedDuration(
			ctx,
			input.Value,
			apiKeyConfig.BlockDuration,
		); saveErr != nil {
			return dto.ApiKeyAllow{}, saveErr
		}
	}

	return dto.ApiKeyAllow{
		Allow: isAllowed,
	}, nil
}
