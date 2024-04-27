package entity

const (
	IPPrefixRateKey          = "rate:ip"
	IPPrefixBlockDurationKey = "block:ip"
	StatusIPBlocked          = "IPBlocked"
)

type IP struct {
	value string

	BlockDuration int64
	RateLimiter   RateLimiter
}

func (ip *IP) SaveValue(ipValue string) {
	ip.value = ipValue
}

func (ip *IP) Value() string {
	return ip.value
}
