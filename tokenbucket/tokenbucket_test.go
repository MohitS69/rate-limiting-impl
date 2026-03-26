package tokenbucket

import (
	"sync"
	"testing"
	"time"
)

func TestBasicAllow(t *testing.T) {
	l := New(5, 1) // capacity=5, refill=1/sec

	// Should allow 5 requests (bucket starts full)
	for i := 0; i < 5; i++ {
		if !l.Allow() {
			t.Fatalf("request %d should have been allowed", i+1)
		}
	}

	// 6th should be rejected
	if l.Allow() {
		t.Fatal("6th request should have been rejected")
	}
}

func TestRefill(t *testing.T) {
	l := New(5, 100) // capacity=5, refill=100/sec (fast refill for testing)

	// Drain all tokens
	for i := 0; i < 5; i++ {
		l.Allow()
	}

	// Wait for refill
	time.Sleep(60 * time.Millisecond)

	// Should have refilled some tokens
	if !l.Allow() {
		t.Fatal("should have refilled at least 1 token after 60ms at 100/sec")
	}
}

func TestAllowN(t *testing.T) {
	l := New(10, 1)

	// Allow 7 tokens at once
	if !l.AllowN(7) {
		t.Fatal("should allow 7 tokens from full bucket of 10")
	}

	// Only 3 left — asking for 5 should fail
	if l.AllowN(5) {
		t.Fatal("should reject 5 tokens when only ~3 remain")
	}

	// Asking for 3 should succeed
	if !l.AllowN(3) {
		t.Fatal("should allow 3 tokens when 3 remain")
	}
}

func TestCapacity(t *testing.T) {
	l := New(10, 1)
	if l.Capacity() != 10 {
		t.Fatalf("expected capacity 10, got %f", l.Capacity())
	}
}

func TestReset(t *testing.T) {
	l := New(5, 1)

	// Drain bucket
	for i := 0; i < 5; i++ {
		l.Allow()
	}
	if l.Allow() {
		t.Fatal("should be empty")
	}

	// Reset and verify
	l.Reset()
	if !l.Allow() {
		t.Fatal("should allow after reset")
	}
}

func TestTokens(t *testing.T) {
	l := New(10, 1)
	l.AllowN(4)
	tokens := l.Tokens()
	if tokens < 5.9 || tokens > 6.1 {
		t.Fatalf("expected ~6 tokens, got %f", tokens)
	}
}

func TestConcurrency(t *testing.T) {
	l := New(1000, 0) // no refill, 1000 capacity
	var wg sync.WaitGroup
	allowed := make(chan bool, 2000)

	for i := 0; i < 2000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			allowed <- l.Allow()
		}()
	}
	wg.Wait()
	close(allowed)

	count := 0
	for a := range allowed {
		if a {
			count++
		}
	}
	if count != 1000 {
		t.Fatalf("expected exactly 1000 allowed, got %d", count)
	}
}
