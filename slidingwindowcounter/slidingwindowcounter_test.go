package slidingwindowcounter

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

	// Should reject when estimated count reaches limit
	if l.Allow() {
		t.Fatal("should be at limit")
	}
}

func TestAllowAt(t *testing.T) {
	l := New(10, 1*time.Minute)

	base := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	// Fill the first window with 8 requests
	for i := 0; i < 8; i++ {
		if !l.AllowAt(base.Add(time.Duration(i) * time.Second)) {
			t.Fatalf("request %d in first window should be allowed", i+1)
		}
	}

	// Move to 30 seconds into the next window
	// Previous window had 8. Overlap ratio = (60-30)/60 = 0.5
	// Estimated = 8 * 0.5 + 0 = 4, so we should have room for 6 more
	nextWindow := base.Add(90 * time.Second) // 12:01:30
	for i := 0; i < 6; i++ {
		if !l.AllowAt(nextWindow.Add(time.Duration(i) * time.Millisecond)) {
			t.Fatalf("request %d in second window should be allowed", i+1)
		}
	}
}

func TestWindowAdvance(t *testing.T) {
	l := New(5, 100*time.Millisecond)

	for i := 0; i < 5; i++ {
		l.Allow()
	}
	if l.Allow() {
		t.Fatal("should be at limit")
	}

	// Wait for window to pass
	time.Sleep(200 * time.Millisecond)

	if !l.Allow() {
		t.Fatal("should allow after windows advance")
	}
}

func TestEstimatedCount(t *testing.T) {
	l := New(100, 1*time.Second)

	l.Allow()
	l.Allow()
	l.Allow()

	est := l.EstimatedCount()
	if est < 2.9 || est > 3.1 {
		t.Fatalf("expected estimated count ~3, got %f", est)
	}
}

func TestRemaining(t *testing.T) {
	l := New(10, 1*time.Second)

	for i := 0; i < 4; i++ {
		l.Allow()
	}

	rem := l.Remaining()
	if rem < 5 || rem > 6 {
		t.Fatalf("expected ~6 remaining, got %d", rem)
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
