package slidingwindowlog

import (
	"sync"
	"testing"
	"time"
)

func TestBasicAllow(t *testing.T) {
	l := New(5, 1*time.Second)

	for i := 0; i < 5; i++ {
		if !l.Allow() {
			t.Fatalf("request %d should have been allowed", i+1)
		}
	}

	if l.Allow() {
		t.Fatal("6th request should have been rejected")
	}
}

func TestEviction(t *testing.T) {
	l := New(3, 100*time.Millisecond)

	for i := 0; i < 3; i++ {
		l.Allow()
	}
	if l.Allow() {
		t.Fatal("should be at limit")
	}

	// Wait for entries to expire
	time.Sleep(150 * time.Millisecond)

	if !l.Allow() {
		t.Fatal("should allow after old entries expire")
	}
}

func TestAllowAt(t *testing.T) {
	l := New(3, 1*time.Minute)

	base := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	// Allow 3 requests
	for i := 0; i < 3; i++ {
		if !l.AllowAt(base.Add(time.Duration(i) * time.Second)) {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}

	// 4th at same time should be rejected
	if l.AllowAt(base.Add(3 * time.Second)) {
		t.Fatal("4th request should be rejected")
	}

	// 61 seconds later, original entries should be evicted
	if !l.AllowAt(base.Add(61 * time.Second)) {
		t.Fatal("should allow after eviction of old entries")
	}
}

func TestCount(t *testing.T) {
	l := New(10, 1*time.Second)

	l.Allow()
	l.Allow()
	l.Allow()

	if c := l.Count(); c != 3 {
		t.Fatalf("expected count 3, got %d", c)
	}
}

func TestRemaining(t *testing.T) {
	l := New(10, 1*time.Second)

	for i := 0; i < 7; i++ {
		l.Allow()
	}

	if r := l.Remaining(); r != 3 {
		t.Fatalf("expected 3 remaining, got %d", r)
	}
}

func TestReset(t *testing.T) {
	l := New(5, 1*time.Second)

	for i := 0; i < 5; i++ {
		l.Allow()
	}

	l.Reset()

	if !l.Allow() {
		t.Fatal("should allow after reset")
	}
}

func TestConcurrency(t *testing.T) {
	l := New(1000, 10*time.Second)
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
