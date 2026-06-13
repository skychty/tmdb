package tmdb

import (
	"context"
	"errors"
	"time"

	"golang.org/x/time/rate"
)

var ErrQueueTimeout = errors.New("tmdb rate limit queue timeout")

type RateLimiter struct {
	limiter      *rate.Limiter
	queueTimeout time.Duration
}

func NewRateLimiter(reqPerSec float64, burst int, queueTimeout time.Duration) *RateLimiter {
	if reqPerSec <= 0 {
		reqPerSec = 40
	}
	if burst <= 0 {
		burst = int(reqPerSec)
	}
	return &RateLimiter{
		limiter:      rate.NewLimiter(rate.Limit(reqPerSec), burst),
		queueTimeout: queueTimeout,
	}
}

func (r *RateLimiter) Acquire(ctx context.Context) error {
	if r == nil {
		return nil
	}

	waitCtx, cancel := context.WithTimeout(ctx, r.queueTimeout)
	defer cancel()

	if err := r.limiter.Wait(waitCtx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return ErrQueueTimeout
		}
		return err
	}
	return nil
}
