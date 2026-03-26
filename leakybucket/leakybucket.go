// Package leakybucket implements the Leaky Bucket rate limiting algorithm.
//
// The leaky bucket works like a queue with a fixed drain rate. Incoming
// requests are added to the bucket (queue). The bucket "leaks" (processes)
// requests at a constant rate. If the bucket is full when a new request
// arrives, the request is rejected.
//
// This produces a perfectly smooth, uniform output rate regardless of
// how bursty the input traffic is.
package leakybucket

import (
	"sync"
	"time"
)

// Limiter is a leaky bucket rate limiter.
type Limiter struct {
	mu sync.Mutex

	capacity  float64   // max number of requests the bucket can hold
	water     float64   // current "water level" (pending requests)
	leakRate  float64   // requests drained per second
	lastCheck time.Time // last time we drained the bucket
}

// New creates a new Leaky Bucket limiter.
//
// capacity: maximum number of pending requests in the bucket.
// leakRate: how many requests are drained (processed) per second.
func New(capacity float64, leakRate float64) *Limiter {
	return &Limiter{
		capacity:  capacity,
		water:     0,
		leakRate:  leakRate,
		lastCheck: time.Now(),
	}
}

// Allow checks if a request can be added to the bucket. If the bucket has
// room after draining, the request is accepted. Otherwise it is rejected.
func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.drain()

	if l.water+1 <= l.capacity {
		l.water++
		return true
	}
	return false
}

// drain removes "leaked" water based on elapsed time.
func (l *Limiter) drain() {
	now := time.Now()
	elapsed := now.Sub(l.lastCheck).Seconds()
	leaked := elapsed * l.leakRate

	l.water -= leaked
	if l.water < 0 {
		l.water = 0
	}
	l.lastCheck = now
}

// WaterLevel returns the current water level (pending requests) in the bucket.
func (l *Limiter) WaterLevel() float64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.drain()
	return l.water
}

// Capacity returns the maximum capacity of the bucket.
func (l *Limiter) Capacity() float64 {
	return l.capacity
}

// Reset empties the bucket.
func (l *Limiter) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.water = 0
	l.lastCheck = time.Now()
}
