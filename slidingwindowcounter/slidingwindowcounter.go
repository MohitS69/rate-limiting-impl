// Package slidingwindowcounter implements the Sliding Window Counter rate
// limiting algorithm.
//
// This is a hybrid of the Fixed Window Counter and the Sliding Window Log.
// It keeps counters for the current and previous fixed windows, then uses
// weighted interpolation to estimate the request count within the sliding window.
//
//	estimated_count = prev_count * overlap_ratio + curr_count
//
// For example, if the window is 1 minute, we are 15 seconds into the current
// minute, and the previous minute had 40 requests while the current has 10:
//
//	estimated = 40 * (45/60) + 10 = 30 + 10 = 40
//
// This provides near-sliding-window accuracy with O(1) memory — only two
// counters are needed instead of a full timestamp log.
package slidingwindowcounter

import (
	"sync"
	"time"
)

// Limiter is a sliding window counter rate limiter.
type Limiter struct {
	mu sync.Mutex

	limit      int           // max requests per window
	windowSize time.Duration // window duration

	currWindowStart time.Time // start of the current fixed window
	currCount       int       // requests in the current fixed window
	prevCount       int       // requests in the previous fixed window
}

// New creates a new Sliding Window Counter limiter.
//
// limit: maximum number of requests allowed per sliding window.
// windowSize: the duration of the window (e.g., 1*time.Minute).
func New(limit int, windowSize time.Duration) *Limiter {
	return &Limiter{
		limit:           limit,
		windowSize:      windowSize,
		currWindowStart: time.Now().Truncate(windowSize),
		currCount:       0,
		prevCount:       0,
	}
}

// Allow checks if a request is allowed within the estimated sliding window.
func (l *Limiter) Allow() bool {
	return l.AllowAt(time.Now())
}

// AllowAt is like Allow but uses a caller-provided timestamp.
func (l *Limiter) AllowAt(now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.advanceWindow(now)

	if l.estimate(now) < float64(l.limit) {
		l.currCount++
		return true
	}
	return false
}

// advanceWindow rolls forward the fixed windows as needed.
func (l *Limiter) advanceWindow(now time.Time) {
	windowStart := now.Truncate(l.windowSize)

	if windowStart == l.currWindowStart {
		return // still in the same window
	}

	// Check how many windows have passed.
	diff := windowStart.Sub(l.currWindowStart)

	if diff == l.windowSize {
		// Exactly one window rotation — previous window becomes the old current.
		l.prevCount = l.currCount
		l.currCount = 0
	} else {
		// More than one window has passed — previous window data is stale.
		l.prevCount = 0
		l.currCount = 0
	}
	l.currWindowStart = windowStart
}

// estimate returns the weighted request count for the sliding window.
func (l *Limiter) estimate(now time.Time) float64 {
	elapsed := now.Sub(l.currWindowStart).Seconds()
	windowSec := l.windowSize.Seconds()

	// Fraction of the previous window that overlaps with our sliding window.
	overlapRatio := (windowSec - elapsed) / windowSec
	if overlapRatio < 0 {
		overlapRatio = 0
	}

	return float64(l.prevCount)*overlapRatio + float64(l.currCount)
}

// EstimatedCount returns the current estimated request count within the
// sliding window.
func (l *Limiter) EstimatedCount() float64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	l.advanceWindow(now)
	return l.estimate(now)
}

// Remaining returns an approximate number of requests still allowed.
func (l *Limiter) Remaining() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	l.advanceWindow(now)
	est := l.estimate(now)
	rem := float64(l.limit) - est
	if rem < 0 {
		return 0
	}
	return int(rem)
}

// Reset clears both window counters.
func (l *Limiter) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.currCount = 0
	l.prevCount = 0
	l.currWindowStart = time.Now().Truncate(l.windowSize)
}

// Limit returns the max requests allowed per window.
func (l *Limiter) Limit() int {
	return l.limit
}

// WindowSize returns the duration of the window.
func (l *Limiter) WindowSize() time.Duration {
	return l.windowSize
}
