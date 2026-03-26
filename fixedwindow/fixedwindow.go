// Package fixedwindow implements the Fixed Window Counter rate limiting algorithm.
//
// The fixed window counter divides time into fixed-duration windows (e.g., 1 minute).
// A counter tracks the number of requests in the current window. If the counter
// exceeds the limit, subsequent requests are rejected until the next window begins.
//
// When the window rolls over, the counter resets to zero.
package fixedwindow

import (
	"sync"
	"time"
)

// Limiter is a fixed window counter rate limiter.
type Limiter struct {
	mu sync.Mutex

	limit      int           // max requests per window
	counter    int           // current request count in this window
	windowSize time.Duration // duration of each fixed window
	windowStart time.Time    // start of the current window
}

// New creates a new Fixed Window Counter limiter.
//
// limit: maximum number of requests allowed per window.
// windowSize: duration of each window (e.g., 1*time.Minute).
func New(limit int, windowSize time.Duration) *Limiter {
	return &Limiter{
		limit:       limit,
		counter:     0,
		windowSize:  windowSize,
		windowStart: time.Now().Truncate(windowSize),
	}
}

// Allow checks if a request is allowed in the current window.
// If the window has rolled over, the counter resets.
func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	windowStart := now.Truncate(l.windowSize)

	// If we've moved to a new window, reset the counter.
	if windowStart != l.windowStart {
		l.windowStart = windowStart
		l.counter = 0
	}

	if l.counter < l.limit {
		l.counter++
		return true
	}
	return false
}

// Count returns the number of requests counted in the current window.
func (l *Limiter) Count() int {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	windowStart := now.Truncate(l.windowSize)

	if windowStart != l.windowStart {
		return 0
	}
	return l.counter
}

// Remaining returns how many more requests are allowed in the current window.
func (l *Limiter) Remaining() int {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	windowStart := now.Truncate(l.windowSize)

	if windowStart != l.windowStart {
		return l.limit
	}
	return l.limit - l.counter
}

// Reset clears the counter for the current window.
func (l *Limiter) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.counter = 0
	l.windowStart = time.Now().Truncate(l.windowSize)
}

// Limit returns the max requests allowed per window.
func (l *Limiter) Limit() int {
	return l.limit
}

// WindowSize returns the duration of each window.
func (l *Limiter) WindowSize() time.Duration {
	return l.windowSize
}
