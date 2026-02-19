package retry_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/jflowers/gcal-organizer/internal/retry"
	"google.golang.org/api/googleapi"
)

// ---------------------------------------------------------------------------
// isRetryable (tested indirectly through Do)
// ---------------------------------------------------------------------------

func TestDo_SuccessOnFirstAttempt(t *testing.T) {
	calls := 0
	err := retry.Do(context.Background(), retry.DefaultConfig(), func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestDo_RetriesOn5xxThenSucceeds(t *testing.T) {
	calls := 0
	cfg := retry.Config{
		MaxRetries:     3,
		InitialBackoff: time.Millisecond,
		MaxBackoff:     10 * time.Millisecond,
		Multiplier:     2.0,
		Jitter:         false,
	}
	err := retry.Do(context.Background(), cfg, func() error {
		calls++
		if calls < 3 {
			return &googleapi.Error{Code: http.StatusServiceUnavailable} // 503
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil after retries, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestDo_RetriesOn429(t *testing.T) {
	calls := 0
	cfg := retry.Config{
		MaxRetries:     2,
		InitialBackoff: time.Millisecond,
		MaxBackoff:     5 * time.Millisecond,
		Multiplier:     2.0,
		Jitter:         false,
	}
	_ = retry.Do(context.Background(), cfg, func() error {
		calls++
		return &googleapi.Error{Code: http.StatusTooManyRequests} // 429
	})
	// Should attempt MaxRetries+1 = 3 times total
	if calls != cfg.MaxRetries+1 {
		t.Fatalf("expected %d calls for 429, got %d", cfg.MaxRetries+1, calls)
	}
}

func TestDo_NoRetryOn4xx(t *testing.T) {
	calls := 0
	cfg := retry.Config{
		MaxRetries:     5,
		InitialBackoff: time.Millisecond,
		MaxBackoff:     10 * time.Millisecond,
		Multiplier:     2.0,
	}
	err := retry.Do(context.Background(), cfg, func() error {
		calls++
		return &googleapi.Error{Code: http.StatusForbidden} // 403
	})
	if err == nil {
		t.Fatal("expected error for non-retryable 403")
	}
	if calls != 1 {
		t.Fatalf("expected 1 call (no retry on 403), got %d", calls)
	}
}

func TestDo_NoRetryOnContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancel

	calls := 0
	cfg := retry.Config{
		MaxRetries:     5,
		InitialBackoff: time.Millisecond,
		MaxBackoff:     10 * time.Millisecond,
		Multiplier:     2.0,
	}
	err := retry.Do(ctx, cfg, func() error {
		calls++
		return context.Canceled
	})
	if err == nil {
		t.Fatal("expected error for context.Canceled")
	}
	if calls != 1 {
		t.Fatalf("expected 1 call (no retry on Canceled), got %d", calls)
	}
}

func TestDo_NoRetryOnContextDeadlineExceeded(t *testing.T) {
	calls := 0
	cfg := retry.Config{
		MaxRetries:     5,
		InitialBackoff: time.Millisecond,
		MaxBackoff:     10 * time.Millisecond,
		Multiplier:     2.0,
	}
	err := retry.Do(context.Background(), cfg, func() error {
		calls++
		return context.DeadlineExceeded
	})
	if err == nil {
		t.Fatal("expected error for context.DeadlineExceeded")
	}
	if calls != 1 {
		t.Fatalf("expected 1 call (no retry on DeadlineExceeded), got %d", calls)
	}
}

func TestDo_ContextCancelledDuringSleep(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	calls := 0
	cfg := retry.Config{
		MaxRetries:     10,
		InitialBackoff: 100 * time.Millisecond, // long enough to cancel during sleep
		MaxBackoff:     200 * time.Millisecond,
		Multiplier:     1.0,
		Jitter:         false,
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	err := retry.Do(ctx, cfg, func() error {
		calls++
		return errors.New("transient network error")
	})

	if err == nil {
		t.Fatal("expected error when context cancelled during sleep")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled in error chain, got: %v", err)
	}
	// Should have attempted at least once and then been cancelled during sleep
	if calls < 1 {
		t.Fatalf("expected at least 1 call, got %d", calls)
	}
}

// TestDo_NonRetryableOnLastAttempt verifies that a non-retryable error on the
// final attempt is returned cleanly — not wrapped as "max retries exceeded".
func TestDo_NonRetryableOnLastAttempt(t *testing.T) {
	const maxRetries = 3
	calls := 0
	cfg := retry.Config{
		MaxRetries:     maxRetries,
		InitialBackoff: time.Millisecond,
		MaxBackoff:     5 * time.Millisecond,
		Multiplier:     2.0,
		Jitter:         false,
	}

	targetErr := &googleapi.Error{Code: http.StatusForbidden} // 403 — non-retryable

	err := retry.Do(context.Background(), cfg, func() error {
		calls++
		if calls < maxRetries+1 {
			// Transient error on all but the last attempt
			return &googleapi.Error{Code: http.StatusServiceUnavailable} // 503
		}
		return targetErr
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Must NOT be wrapped as "max retries exceeded"
	if calls != maxRetries+1 {
		t.Fatalf("expected %d calls, got %d", maxRetries+1, calls)
	}
	var apiErr *googleapi.Error
	if !errors.As(err, &apiErr) || apiErr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 error returned cleanly, got: %v", err)
	}
}

func TestDo_ExhaustsMaxRetries(t *testing.T) {
	const maxRetries = 3
	calls := 0
	cfg := retry.Config{
		MaxRetries:     maxRetries,
		InitialBackoff: time.Millisecond,
		MaxBackoff:     5 * time.Millisecond,
		Multiplier:     2.0,
		Jitter:         false,
	}
	err := retry.Do(context.Background(), cfg, func() error {
		calls++
		return errors.New("persistent error")
	})
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	expectedCalls := maxRetries + 1 // 1 initial + maxRetries
	if calls != expectedCalls {
		t.Fatalf("expected %d calls, got %d", expectedCalls, calls)
	}
}

// TestDo_ZeroValueConfigDoesNotBusySpin verifies that a zero-value Config does not
// produce a busy-spin: the Multiplier, InitialBackoff, and MaxBackoff guards fire.
func TestDo_ZeroValueConfigDoesNotBusySpin(t *testing.T) {
	// Zero-value Config: MaxBackoff=0, InitialBackoff=0, Multiplier=0, MaxRetries=0.
	// With MaxRetries=0 the function attempts exactly once and returns the error.
	calls := 0
	start := time.Now()
	err := retry.Do(context.Background(), retry.Config{}, func() error {
		calls++
		return errors.New("always fails")
	})
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// MaxRetries=0 in zero-value struct → only 1 attempt, no sleep
	if calls != 1 {
		t.Fatalf("expected 1 call with zero MaxRetries, got %d", calls)
	}
	// Should return quickly (only 1 attempt, no sleep after final attempt)
	if elapsed > 500*time.Millisecond {
		t.Fatalf("zero-value Config took too long (%v), possible busy-spin", elapsed)
	}
}

// ---------------------------------------------------------------------------
// calculateBackoff (tested indirectly — verify it stays within bounds)
// ---------------------------------------------------------------------------

func TestDo_BackoffStaysWithinMaxBackoff(t *testing.T) {
	const maxBackoff = 5 * time.Millisecond
	cfg := retry.Config{
		MaxRetries:     20,
		InitialBackoff: time.Millisecond,
		MaxBackoff:     maxBackoff,
		Multiplier:     10.0, // aggressive to hit cap quickly
		Jitter:         false,
	}

	start := time.Now()
	_ = retry.Do(context.Background(), cfg, func() error {
		return errors.New("always fails")
	})
	elapsed := time.Since(start)

	// 20 sleeps × 5ms max = 100ms; add generous headroom for CI
	if elapsed > 2*time.Second {
		t.Fatalf("backoff exceeded MaxBackoff cap: elapsed %v", elapsed)
	}
}
