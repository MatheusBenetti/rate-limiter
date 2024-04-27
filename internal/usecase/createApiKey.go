package usecase

import (
	"context"
	"log"

	"github.com/MatheusBenetti/rate-limiter/internal/dto"
	"github.com/MatheusBenetti/rate-limiter/internal/entity"
)

type CreateApiKeyUseCase struct {
	apiKeyRepository entity.ApiKeyRepository
}

func NewCreateAPIKeyUseCase(apiKeyRepository entity.ApiKeyRepository) *CreateApiKeyUseCase {
	return &CreateApiKeyUseCase{apiKeyRepository: apiKeyRepository}
}

func (cr *CreateApiKeyUseCase) Execute(ctx context.Context, input dto.Input) (dto.Output, error) {
	apiKey := entity.ApiKey{
		BlockDuration: input.BlockDuration,
		RateLimiter: entity.RateLimiter{
			TimeWindow: input.TimeWindow,
			MaxReq:     input.MaxReq,
		},
	}

	if err := apiKey.GenerateValue(); err != nil {
		log.Printf("Error on CreateAPIKeyUseCase generating key value: %s\n", err.Error())
		return dto.Output{}, err
	}

	keyValue, saveErr := cr.apiKeyRepository.Save(ctx, &apiKey)
	if saveErr != nil {
		log.Printf("Error on CreateAPIKeyUseCase saving data: %s\n", saveErr.Error())
		return dto.Output{}, saveErr
	}

	log.Println("Saved with success")
	return dto.Output{
		Api_Key: keyValue,
	}, nil
}
