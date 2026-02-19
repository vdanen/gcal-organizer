// Package retry provides exponential backoff retry logic for API calls.
package retry

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

	"google.golang.org/api/googleapi"
)

// Config holds retry configuration.
type Config struct {
	MaxRetries     int           // Maximum number of retry attempts (default: 5)
	InitialBackoff time.Duration // Initial backoff duration (default: 1s)
	MaxBackoff     time.Duration // Maximum backoff duration (default: 32s)
	Multiplier     float64       // Backoff multiplier (default: 2.0)
	Jitter         bool          // Add random jitter to backoff (default: true)
}

// DefaultConfig returns sensible defaults for Google API retry logic.
func DefaultConfig() Config {
	return Config{
		MaxRetries:     5,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     32 * time.Second,
		Multiplier:     2.0,
		Jitter:         true,
	}
}

// Do executes fn with exponential backoff retry on transient errors
// (HTTP 5xx, 429 rate limit, network errors). Non-retryable errors (4xx
// except 429) are returned immediately without retrying.
//
// If cfg.MaxBackoff is zero, Do falls back to DefaultConfig().MaxBackoff to
// prevent a busy-spin where every sleep is zero duration.
func Do(ctx context.Context, cfg Config, fn func() error) error {
	// Guard against zero-value Config fields that would produce a busy-spin:
	//   MaxBackoff=0    → every sleep is 0 ns
	//   InitialBackoff=0→ every sleep is 0 ns
	//   Multiplier=0    → math.Pow(0, attempt)=0 for attempt≥1, also 0 ns sleeps
	if cfg.MaxBackoff == 0 {
		cfg.MaxBackoff = DefaultConfig().MaxBackoff
	}
	if cfg.InitialBackoff == 0 {
		cfg.InitialBackoff = DefaultConfig().InitialBackoff
	}
	if cfg.Multiplier == 0 {
		cfg.Multiplier = DefaultConfig().Multiplier
	}

	var lastErr error

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
		}

		// Non-transient errors bail out immediately regardless of which attempt
		// number produced them. This check runs before the final-attempt break
		// so that a non-retryable error on the last attempt is still returned
		// cleanly (not wrapped in "max retries exceeded").
		if !isRetryable(lastErr) {
			return lastErr
		}

		// Don't sleep after the final attempt.
		if attempt == cfg.MaxRetries {
			break
		}

		backoff := calculateBackoff(cfg, attempt)
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		case <-time.After(backoff):
		}
	}

	return fmt.Errorf("max retries exceeded (%d attempts): %w", cfg.MaxRetries+1, lastErr)
}

// isRetryable reports whether err is a transient error worth retrying.
func isRetryable(err error) bool {
	// Both Canceled and DeadlineExceeded mean the caller's context has expired;
	// retrying would either immediately fail again or silently overshoot the
	// caller-imposed deadline, so neither is retryable.
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	var apiErr *googleapi.Error
	if errors.As(err, &apiErr) {
		// Retry server errors (5xx) and rate limiting (429).
		return apiErr.Code >= 500 || apiErr.Code == 429
	}

	// For all other errors (network timeouts, connection resets, etc.)
	// assume transient and allow a retry.
	return true
}

// calculateBackoff returns the sleep duration for the given attempt number.
func calculateBackoff(cfg Config, attempt int) time.Duration {
	backoff := float64(cfg.InitialBackoff) * math.Pow(cfg.Multiplier, float64(attempt))
	if backoff > float64(cfg.MaxBackoff) {
		backoff = float64(cfg.MaxBackoff)
	}
	if cfg.Jitter {
		// Full-range jitter (0–100 % of backoff) distributes retrying clients
		// evenly across the entire interval, preventing thundering herd.
		backoff = rand.Float64() * backoff
	}
	return time.Duration(backoff)
}
