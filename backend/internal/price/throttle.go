package price

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// RateBudget caps how many calls are allowed per fixed window, shared across
// processes via Redis. Acquire blocks until a slot is available.
type RateBudget interface {
	Acquire(ctx context.Context) error
}

// NoopBudget never blocks.
type NoopBudget struct{}

// Acquire implements RateBudget.
func (NoopBudget) Acquire(ctx context.Context) error { return nil }

const rateLimitScript = `
local c = redis.call('INCR', KEYS[1])
if c == 1 then redis.call('EXPIRE', KEYS[1], ARGV[2]) end
return c
`

// RedisBudget implements a fixed-window rate limit in Redis. On Redis errors it
// degrades to unlimited and warns once.
type RedisBudget struct {
	rdb    *redis.Client
	limit  int64
	window time.Duration
	warned sync.Once
}

// NewRedisBudget returns a RateBudget backed by Redis keyed on a fixed window.
func NewRedisBudget(rdb *redis.Client, limit int64, window time.Duration) RateBudget {
	return &RedisBudget{rdb: rdb, limit: limit, window: window}
}

// Acquire implements RateBudget.
func (b *RedisBudget) Acquire(ctx context.Context) error {
	windowSecs := int64(b.window.Seconds())
	if windowSecs < 1 {
		windowSecs = 1
	}
	for {
		key := "yahoo:rate:" + strconv.FormatInt(time.Now().Unix(), 10)
		count, err := b.rdb.Eval(ctx, rateLimitScript, []string{key}, b.limit, windowSecs).Int64()
		if err != nil {
			b.warned.Do(func() {
				log.Warn().Err(err).Msg("redis rate budget unavailable, continuing without global rate limit")
			})
			return nil
		}
		if count <= b.limit {
			return nil
		}
		// Wait for the current window to roll over before retrying.
		wait := time.Until(time.Now().Truncate(time.Second).Add(time.Second)) + 10*time.Millisecond
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

type throttleJob struct {
	ctx context.Context
	fn  func(ctx context.Context) error
	res chan error
}

// throttler serializes Yahoo HTTP calls through a FIFO queue with a minimum gap
// between calls and an optional global rate budget.
type throttler struct {
	queue       chan *throttleJob
	minInterval time.Duration
	budget      RateBudget
}

func newThrottler(minInterval time.Duration, budget RateBudget) *throttler {
	if budget == nil {
		budget = NoopBudget{}
	}
	t := &throttler{
		queue:       make(chan *throttleJob, 256),
		minInterval: minInterval,
		budget:      budget,
	}
	go t.run()
	return t
}

// Do queues fn and blocks until it has been executed (or the context is done).
func (t *throttler) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	job := &throttleJob{ctx: ctx, fn: fn, res: make(chan error, 1)}
	select {
	case t.queue <- job:
	case <-ctx.Done():
		return ctx.Err()
	}
	select {
	case err := <-job.res:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (t *throttler) run() {
	var last time.Time
	for job := range t.queue {
		if job.ctx.Err() != nil {
			job.res <- job.ctx.Err()
			continue
		}
		if !last.IsZero() {
			if wait := time.Until(last.Add(t.minInterval)); wait > 0 {
				select {
				case <-time.After(wait):
				case <-job.ctx.Done():
					job.res <- job.ctx.Err()
					continue
				}
			}
		}
		if err := t.budget.Acquire(job.ctx); err != nil {
			job.res <- err
			continue
		}
		err := job.fn(job.ctx)
		last = time.Now()
		job.res <- err
	}
}
