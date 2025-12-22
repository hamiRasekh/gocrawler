package utils

import (
	"context"

	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiter *rate.Limiter
}

func NewRateLimiter(requestsPerSecond int) *RateLimiter {
	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Limit(requestsPerSecond), requestsPerSecond),
	}
}

func (rl *RateLimiter) Wait(ctx context.Context) error {
	return rl.limiter.Wait(ctx)
}

func (rl *RateLimiter) Allow() bool {
	return rl.limiter.Allow()
}

func (rl *RateLimiter) SetRate(requestsPerSecond int) {
	rl.limiter.SetLimit(rate.Limit(requestsPerSecond))
	rl.limiter.SetBurst(requestsPerSecond)
}

type DomainRateLimiter struct {
	limiters map[string]*RateLimiter
	global   *RateLimiter
}

func NewDomainRateLimiter(globalRPS int) *DomainRateLimiter {
	return &DomainRateLimiter{
		limiters: make(map[string]*RateLimiter),
		global:   NewRateLimiter(globalRPS),
	}
}

func (drl *DomainRateLimiter) WaitForDomain(ctx context.Context, domain string) error {
	// Wait for global limiter
	if err := drl.global.Wait(ctx); err != nil {
		return err
	}
	
	// Wait for domain-specific limiter if exists
	if limiter, exists := drl.limiters[domain]; exists {
		return limiter.Wait(ctx)
	}
	
	return nil
}

func (drl *DomainRateLimiter) SetDomainRate(domain string, requestsPerSecond int) {
	drl.limiters[domain] = NewRateLimiter(requestsPerSecond)
}

