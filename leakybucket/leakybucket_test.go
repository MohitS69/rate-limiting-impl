package leakybucket

import (
	"sync"
	"testing"
	"time"
)

func TestBasicAllow(t *testing.T) {
	l := New(5, 1) // capacity=5, leakRate=1/sec

	// Should allow 5 requests
	for i := 0; i < 5; i++ {
		if !l.Allow() {
			t.Fatalf("request %d should have been allowed", i+1)
		}
	}

	// 6th should be rejected (bucket full)
	if l.Allow() {
		t.Fatal("6th request should have been rejected")
	}
}

func TestDrain(t *testing.T) {
	l := New(5, 100) // capacity=5, leakRate=100/sec (fast drain)

	// Fill the bucket
	for i := 0; i < 5; i++ {
		l.Allow()
	}

	// Wait for some draining
	time.Sleep(60 * time.Millisecond)

	// Should have room now
	if !l.Allow() {
		t.Fatal("should have drained enough for at least 1 request after 60ms at 100/sec")
	}
}

func TestWaterLevel(t *testing.T) {
	l := New(10, 1)

	l.Allow()
	l.Allow()
	l.Allow()

	level := l.WaterLevel()
	if level < 2.9 || level > 3.1 {
		t.Fatalf("expected water level ~3, got %f", level)
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

	for i := 0; i < 5; i++ {
		l.Allow()
	}
	if l.Allow() {
		t.Fatal("should be full")
	}

	l.Reset()
	if !l.Allow() {
		t.Fatal("should allow after reset")
	}
}

func TestConcurrency(t *testing.T) {
	l := New(1000, 0) // no drain, 1000 capacity
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
