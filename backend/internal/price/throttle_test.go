package price

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestThrottlerExecutesFn(t *testing.T) {
	th := newThrottler(time.Millisecond, NoopBudget{})
	var calls atomic.Int32
	if err := th.Do(context.Background(), func(ctx context.Context) error {
		calls.Add(1)
		return nil
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("expected 1 call, got %d", calls.Load())
	}
}

func TestThrottlerMinInterval(t *testing.T) {
	const gap = 50 * time.Millisecond
	th := newThrottler(gap, NoopBudget{})
	var calls atomic.Int32
	start := time.Now()
	for i := 0; i < 3; i++ {
		if err := th.Do(context.Background(), func(ctx context.Context) error {
			calls.Add(1)
			return nil
		}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if calls.Load() != 3 {
		t.Fatalf("expected 3 calls, got %d", calls.Load())
	}
	// Two gaps between three calls; allow generous tolerance for CI noise.
	if elapsed := time.Since(start); elapsed < 2*(gap/2) {
		t.Fatalf("expected at least %v of pacing, elapsed %v", 2*(gap/2), elapsed)
	}
}

func TestThrottlerSkipsCancelledCtx(t *testing.T) {
	th := newThrottler(time.Millisecond, NoopBudget{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := th.Do(ctx, func(ctx context.Context) error {
		t.Fatal("fn must not run for a cancelled context")
		return nil
	}); err == nil {
		t.Fatal("expected an error for a cancelled context")
	}
}

func TestThrottlerCancellationWhileQueued(t *testing.T) {
	th := newThrottler(time.Millisecond, NoopBudget{})
	release := make(chan struct{})
	fn1Started := make(chan struct{})
	go th.Do(context.Background(), func(ctx context.Context) error {
		close(fn1Started)
		<-release
		return nil
	})
	// Wait until the worker is blocked inside the first fn before queuing a
	// second job, so the second one is guaranteed to be pending on cancel.
	<-fn1Started

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- th.Do(ctx, func(ctx context.Context) error { return nil })
	}()
	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected a cancellation error")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Do did not return after cancellation")
	}
	close(release)
}
