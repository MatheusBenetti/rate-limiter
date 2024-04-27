package usecase

import (
	"context"
	"log"

	"github.com/MatheusBenetti/rate-limiter/config"
	"github.com/MatheusBenetti/rate-limiter/internal/dto"
	"github.com/MatheusBenetti/rate-limiter/internal/entity"
)

type RegisterIP struct {
	ipRepository entity.IPRepository
	config       *config.Config
}

func NewRegisterIPUseCase(
	ipRepository entity.IPRepository,
	config *config.Config,
) *RegisterIP {
	return &RegisterIP{
		ipRepository: ipRepository,
		config:       config,
	}
}

func (ipr *RegisterIP) Execute(
	ctx context.Context,
	input dto.IpReq,
) (dto.IpAllow, error) {
	status, blockedErr := ipr.ipRepository.GetBlockedDuration(ctx, input.IP)
	if blockedErr != nil {
		return dto.IpAllow{}, blockedErr
	}

	if status == entity.StatusIPBlocked {
		log.Println("ip is blocked due to exceeding the maximum number of requests")
		return dto.IpAllow{}, entity.ErrIpAmountReq
	}

	getReq, getReqErr := ipr.ipRepository.GetRequest(ctx, input.IP)
	if getReqErr != nil {
		log.Printf("Error getting IP requests: %s \n", getReqErr.Error())
		return dto.IpAllow{}, getReqErr
	}

	getReq.TimeWindow = ipr.config.RateLimiter.ByIp.TimeWindow
	getReq.MaxReq = ipr.config.RateLimiter.ByIp.MaxReq
	if valErr := getReq.Validate(); valErr != nil {
		log.Printf("Error validation in rate limiter: %s \n", valErr.Error())
		return dto.IpAllow{}, valErr
	}

	getReq.AddReq(input.TimeAdded)
	isAllowed := getReq.Allow(input.TimeAdded)
	if upsertErr := ipr.ipRepository.UpsertRequest(ctx, input.IP, getReq); upsertErr != nil {
		log.Printf("Error updating/inserting rate limit: %s \n", upsertErr.Error())
		return dto.IpAllow{}, upsertErr
	}

	if !isAllowed {
		if saveErr := ipr.ipRepository.SaveBlockedDuration(
			ctx,
			input.IP,
			ipr.config.RateLimiter.ByIp.BlockDuration,
		); saveErr != nil {
			return dto.IpAllow{}, saveErr
		}
	}

	return dto.IpAllow{
		Allow: isAllowed,
	}, nil
}
