// Command ratelimiter-demo demonstrates all five rate limiting algorithms.
package main

import (
	"fmt"
	"time"

	"github.com/mohit/ratelimiter/fixedwindow"
	"github.com/mohit/ratelimiter/leakybucket"
	"github.com/mohit/ratelimiter/slidingwindowcounter"
	"github.com/mohit/ratelimiter/slidingwindowlog"
	"github.com/mohit/ratelimiter/tokenbucket"
)

func main() {
	fmt.Println("═══════════════════════════════════════════════")
	fmt.Println("        Rate Limiting Algorithms Demo")
	fmt.Println("═══════════════════════════════════════════════")

	demoTokenBucket()
	demoLeakyBucket()
	demoFixedWindow()
	demoSlidingWindowLog()
	demoSlidingWindowCounter()
}

func demoTokenBucket() {
	fmt.Println("\n┌─────────────────────────────────────────────┐")
	fmt.Println("│           1. TOKEN BUCKET                    │")
	fmt.Println("└─────────────────────────────────────────────┘")

	// 5 tokens capacity, refill at 2 tokens/sec
	limiter := tokenbucket.New(5, 2)

	fmt.Printf("  Capacity: %.0f, Refill Rate: 2/sec\n\n", limiter.Capacity())

	// Burst: try 7 requests rapidly
	fmt.Println("  Sending 7 rapid requests (burst):")
	for i := 1; i <= 7; i++ {
		result := limiter.Allow()
		status := "✓ ALLOWED"
		if !result {
			status = "✗ REJECTED"
		}
		fmt.Printf("    Request %d: %s (tokens: %.1f)\n", i, status, limiter.Tokens())
	}

	// Wait for refill
	fmt.Println("\n  Waiting 1 second for refill...")
	time.Sleep(1 * time.Second)
	fmt.Printf("  Tokens after 1s: %.1f\n", limiter.Tokens())

	if limiter.Allow() {
		fmt.Println("  Request after wait: ✓ ALLOWED")
	}
}

func demoLeakyBucket() {
	fmt.Println("\n┌─────────────────────────────────────────────┐")
	fmt.Println("│           2. LEAKY BUCKET                    │")
	fmt.Println("└─────────────────────────────────────────────┘")

	// capacity=5, drain at 2 requests/sec
	limiter := leakybucket.New(5, 2)

	fmt.Printf("  Capacity: %.0f, Leak Rate: 2/sec\n\n", limiter.Capacity())

	fmt.Println("  Sending 7 rapid requests:")
	for i := 1; i <= 7; i++ {
		result := limiter.Allow()
		status := "✓ ALLOWED"
		if !result {
			status = "✗ REJECTED"
		}
		fmt.Printf("    Request %d: %s (water: %.1f)\n", i, status, limiter.WaterLevel())
	}

	fmt.Println("\n  Waiting 1 second for drain...")
	time.Sleep(1 * time.Second)
	fmt.Printf("  Water level after 1s: %.1f\n", limiter.WaterLevel())

	if limiter.Allow() {
		fmt.Println("  Request after wait: ✓ ALLOWED")
	}
}

func demoFixedWindow() {
	fmt.Println("\n┌─────────────────────────────────────────────┐")
	fmt.Println("│        3. FIXED WINDOW COUNTER               │")
	fmt.Println("└─────────────────────────────────────────────┘")

	limiter := fixedwindow.New(5, 1*time.Second)

	fmt.Printf("  Limit: %d/window, Window: %v\n\n", limiter.Limit(), limiter.WindowSize())

	fmt.Println("  Sending 7 rapid requests:")
	for i := 1; i <= 7; i++ {
		result := limiter.Allow()
		status := "✓ ALLOWED"
		if !result {
			status = "✗ REJECTED"
		}
		fmt.Printf("    Request %d: %s (count: %d, remaining: %d)\n",
			i, status, limiter.Count(), limiter.Remaining())
	}

	fmt.Println("\n  Waiting for window reset...")
	time.Sleep(1100 * time.Millisecond)
	fmt.Printf("  Count after reset: %d, Remaining: %d\n", limiter.Count(), limiter.Remaining())

	if limiter.Allow() {
		fmt.Println("  Request after reset: ✓ ALLOWED")
	}
}

func demoSlidingWindowLog() {
	fmt.Println("\n┌─────────────────────────────────────────────┐")
	fmt.Println("│       4. SLIDING WINDOW LOG                  │")
	fmt.Println("└─────────────────────────────────────────────┘")

	limiter := slidingwindowlog.New(5, 1*time.Second)

	fmt.Printf("  Limit: %d/window, Window: %v\n\n", limiter.Limit(), limiter.WindowSize())

	fmt.Println("  Sending 7 rapid requests:")
	for i := 1; i <= 7; i++ {
		result := limiter.Allow()
		status := "✓ ALLOWED"
		if !result {
			status = "✗ REJECTED"
		}
		fmt.Printf("    Request %d: %s (log count: %d, remaining: %d)\n",
			i, status, limiter.Count(), limiter.Remaining())
	}

	fmt.Println("\n  Waiting for entries to expire...")
	time.Sleep(1100 * time.Millisecond)
	fmt.Printf("  Log count after expiry: %d, Remaining: %d\n", limiter.Count(), limiter.Remaining())

	if limiter.Allow() {
		fmt.Println("  Request after expiry: ✓ ALLOWED")
	}
}

func demoSlidingWindowCounter() {
	fmt.Println("\n┌─────────────────────────────────────────────┐")
	fmt.Println("│     5. SLIDING WINDOW COUNTER                │")
	fmt.Println("└─────────────────────────────────────────────┘")

	limiter := slidingwindowcounter.New(5, 1*time.Second)

	fmt.Printf("  Limit: %d/window, Window: %v\n\n", limiter.Limit(), limiter.WindowSize())

	fmt.Println("  Sending 7 rapid requests:")
	for i := 1; i <= 7; i++ {
		result := limiter.Allow()
		status := "✓ ALLOWED"
		if !result {
			status = "✗ REJECTED"
		}
		fmt.Printf("    Request %d: %s (estimated: %.1f, remaining: %d)\n",
			i, status, limiter.EstimatedCount(), limiter.Remaining())
	}

	fmt.Println("\n  Waiting for window to advance...")
	time.Sleep(1100 * time.Millisecond)
	fmt.Printf("  Estimated count after advance: %.1f, Remaining: %d\n",
		limiter.EstimatedCount(), limiter.Remaining())

	if limiter.Allow() {
		fmt.Println("  Request after advance: ✓ ALLOWED")
	}

	fmt.Println("\n═══════════════════════════════════════════════")
	fmt.Println("                  Done!")
	fmt.Println("═══════════════════════════════════════════════")
}
