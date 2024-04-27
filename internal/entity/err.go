package entity

import "errors"

var (
	ErrIpAmountReq       = errors.New("you have reached the maximum number of Requests or actions by ip allowed within a certain time frame - blocked")
	ErrApiKeyAmountReq   = errors.New("you have reached the maximum number of Requests or actions by api key allowed within a certain time frame - blocked")
	ErrBlockTimeDuration = errors.New("blocked time duration should be greater than zero")
	ErrTimeWindow        = errors.New("rate limiter time window duration should be greater than zero")
	ErrRateLimiterMaxReq = errors.New("rate limiter maximum requests should be greater than zero")
)
