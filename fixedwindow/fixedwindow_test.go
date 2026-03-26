package fixedwindow

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

func TestWindowReset(t *testing.T) {
	l := New(3, 100*time.Millisecond)

	for i := 0; i < 3; i++ {
		l.Allow()
	}
	if l.Allow() {
		t.Fatal("should be at limit")
	}

	// Wait for window to roll over
	time.Sleep(150 * time.Millisecond)

	if !l.Allow() {
		t.Fatal("should allow after window reset")
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

	for i := 0; i < 4; i++ {
		l.Allow()
	}

	if r := l.Remaining(); r != 6 {
		t.Fatalf("expected 6 remaining, got %d", r)
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

func TestLimit(t *testing.T) {
	l := New(42, 1*time.Second)
	if l.Limit() != 42 {
		t.Fatalf("expected limit 42, got %d", l.Limit())
	}
}

func TestWindowSize(t *testing.T) {
	l := New(10, 5*time.Minute)
	if l.WindowSize() != 5*time.Minute {
		t.Fatalf("expected window 5m, got %v", l.WindowSize())
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
