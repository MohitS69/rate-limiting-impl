// Package slidingwindowlog implements the Sliding Window Log rate limiting algorithm.
//
// This algorithm maintains a sorted log (list) of timestamps for each request.
// When a new request arrives, all timestamps older than the window duration are
// removed. If the remaining count is under the limit, the request is allowed and
// its timestamp is added.
//
// This provides the most accurate rate limiting of all windowed approaches,
// at the cost of higher memory usage (one timestamp stored per request).
package slidingwindowlog

import (
	"sync"
	"time"
)

// Limiter is a sliding window log rate limiter.
type Limiter struct {
	mu sync.Mutex

	limit      int           // max requests per window
	windowSize time.Duration // sliding window duration
	log        []time.Time   // sorted timestamps of requests within the window
}

// New creates a new Sliding Window Log limiter.
//
// limit: maximum number of requests allowed within the sliding window.
// windowSize: the duration of the sliding window (e.g., 1*time.Minute).
func New(limit int, windowSize time.Duration) *Limiter {
	return &Limiter{
		limit:      limit,
		windowSize: windowSize,
		log:        make([]time.Time, 0, limit),
	}
}

// Allow checks if a request is allowed. It evicts expired timestamps,
// then checks whether the log length is under the limit.
func (l *Limiter) Allow() bool {
	return l.AllowAt(time.Now())
}

// AllowAt is like Allow but uses a caller-provided timestamp.
// Useful for testing and replay scenarios.
func (l *Limiter) AllowAt(now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.evict(now)

	if len(l.log) < l.limit {
		l.log = append(l.log, now)
		return true
	}
	return false
}

// evict removes all timestamps that have fallen outside the sliding window.
func (l *Limiter) evict(now time.Time) {
	cutoff := now.Add(-l.windowSize)

	// Binary search for the first valid entry (timestamps are sorted).
	i := 0
	for i < len(l.log) && l.log[i].Before(cutoff) {
		i++
	}

	if i > 0 {
		// Shift remaining entries to the front.
		copy(l.log, l.log[i:])
		l.log = l.log[:len(l.log)-i]
	}
}

// Count returns the number of requests currently in the sliding window.
func (l *Limiter) Count() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.evict(time.Now())
	return len(l.log)
}

// Remaining returns how many more requests are allowed in the current window.
func (l *Limiter) Remaining() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.evict(time.Now())
	return l.limit - len(l.log)
}

// Reset clears the entire log.
func (l *Limiter) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.log = l.log[:0]
}

// Limit returns the max requests allowed per window.
func (l *Limiter) Limit() int {
	return l.limit
}

// WindowSize returns the duration of the sliding window.
func (l *Limiter) WindowSize() time.Duration {
	return l.windowSize
}
