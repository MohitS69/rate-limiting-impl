// Package tokenbucket implements the Token Bucket rate limiting algorithm.
//
// The token bucket algorithm works by maintaining a bucket that holds tokens.
// Tokens are added at a fixed rate (refill rate). Each request consumes one
// or more tokens. If the bucket has enough tokens, the request is allowed;
// otherwise, it is rejected.
package tokenbucket

import (
	"sync"
	"time"
)

// Limiter is a token bucket rate limiter.
type Limiter struct {
	mu sync.Mutex

	capacity   float64   // maximum number of tokens the bucket can hold
	tokens     float64   // current number of tokens in the bucket
	refillRate float64   // tokens added per second
	lastRefill time.Time // last time tokens were refilled
}

// New creates a new Token Bucket limiter.
//
// capacity: the maximum number of tokens. This also controls burst size.
// refillRate: how many tokens are added per second.
func New(capacity float64, refillRate float64) *Limiter {
	return &Limiter{
		capacity:   capacity,
		tokens:     capacity, // start full
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow checks if a single request is allowed and consumes one token if so.
func (l *Limiter) Allow() bool {
	return l.AllowN(1)
}

// AllowN checks if n tokens are available and consumes them if so.
// This enables variable-cost requests (e.g., a bulk API call costs more tokens).
func (l *Limiter) AllowN(n float64) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.refill()

	if l.tokens >= n {
		l.tokens -= n
		return true
	}
	return false
}

// refill adds tokens based on elapsed time since the last refill.
// Tokens are capped at the bucket's capacity.
func (l *Limiter) refill() {
	now := time.Now()
	elapsed := now.Sub(l.lastRefill).Seconds()
	l.tokens += elapsed * l.refillRate
	if l.tokens > l.capacity {
		l.tokens = l.capacity
	}
	l.lastRefill = now
}

// Tokens returns the current (approximate) number of tokens in the bucket.
func (l *Limiter) Tokens() float64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.refill()
	return l.tokens
}

// Capacity returns the maximum capacity of the bucket.
func (l *Limiter) Capacity() float64 {
	return l.capacity
}

// Reset fills the bucket back to full capacity.
func (l *Limiter) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.tokens = l.capacity
	l.lastRefill = time.Now()
}
