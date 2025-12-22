package utils

import (
	"context"
	"math"
	"time"

	"go.uber.org/zap"
)

type RetryConfig struct {
	MaxAttempts      int
	BackoffMultiplier int
	InitialDelay     time.Duration
	MaxDelay         time.Duration
}

type RetryableFunc func() error

func Retry(ctx context.Context, config RetryConfig, fn RetryableFunc) error {
	logger := GetLogger()
	
	var lastErr error
	delay := config.InitialDelay
	
	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		err := fn()
		if err == nil {
			if attempt > 1 {
				logger.Info("Retry succeeded",
					zap.Int("attempt", attempt),
					zap.Int("max_attempts", config.MaxAttempts),
				)
			}
			return nil
		}
		
		lastErr = err
		
		if attempt < config.MaxAttempts {
			// Exponential backoff
			nextDelay := time.Duration(float64(delay) * math.Pow(float64(config.BackoffMultiplier), float64(attempt-1)))
			if nextDelay > config.MaxDelay {
				nextDelay = config.MaxDelay
			}
			
			logger.Warn("Retry attempt failed, retrying",
				zap.Int("attempt", attempt),
				zap.Int("max_attempts", config.MaxAttempts),
				zap.Duration("next_delay", nextDelay),
				zap.Error(err),
			)
			
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(nextDelay):
			}
		}
	}
	
	logger.Error("All retry attempts failed",
		zap.Int("max_attempts", config.MaxAttempts),
		zap.Error(lastErr),
	)
	
	return lastErr
}

