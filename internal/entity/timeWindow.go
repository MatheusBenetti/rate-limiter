package entity

import (
	"sync"
	"time"
)

type RateLimiter struct {
	Req        []time.Time
	TimeWindow int64
	MaxReq     int
	lock       sync.Mutex
}

func (rl *RateLimiter) Allow(fromTime time.Time) bool {
	rl.lock.Lock()
	defer rl.lock.Unlock()

	rl.removeOldReq(fromTime)
	return len(rl.Req) <= rl.MaxReq
}

func (rl *RateLimiter) GetDurationTimeWindow() time.Duration {
	return time.Duration(rl.TimeWindow) * time.Second
}

func (rl *RateLimiter) removeOldReq(fromTime time.Time) {
	threshold := fromTime.Add(-rl.GetDurationTimeWindow())
	start := 0
	for i, t := range rl.Req {
		if t.After(threshold) {
			start = i
			break
		}
	}
	rl.Req = rl.Req[start:]
}

func (rl *RateLimiter) AddReq(request time.Time) {
	rl.Req = append(rl.Req, request)
}

func (rl *RateLimiter) Validate() error {
	if rl.MaxReq == 0 {
		return ErrRateLimiterMaxReq
	}

	if rl.TimeWindow == 0 {
		return ErrTimeWindow
	}

	return nil
}
