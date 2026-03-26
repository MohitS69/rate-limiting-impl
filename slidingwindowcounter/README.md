# Sliding Window Counter Algorithm

## How It Works

The Sliding Window Counter is a **hybrid** of the Fixed Window Counter and the Sliding Window Log. It maintains counters for the **current** and **previous** fixed windows, then uses **weighted interpolation** to estimate the request count within a sliding window.

```
  Previous Window         Current Window
  (12:00 - 12:01)         (12:01 - 12:02)
┌─────────────────────┐ ┌─────────────────────┐
│  prevCount = 40      │ │  currCount = 10      │
└─────────────────────┘ └─────────▲───────────┘
                                  │
                          now = 12:01:15
                          (15s into current window)

  Overlap of previous window = (60 - 15) / 60 = 75%

  Estimated count = 40 × 0.75 + 10 = 30 + 10 = 40
  
  If limit = 50 → ALLOW (40 < 50)
```

### The Formula

```
estimated = prevCount × ((windowSize - elapsed) / windowSize) + currCount
```

Where:
- `prevCount` = total requests in the previous fixed window
- `currCount` = requests so far in the current fixed window
- `elapsed` = time since the current window started
- `windowSize` = duration of each window

### Key Parameters
| Parameter    | Description                              |
|--------------|------------------------------------------|
| `limit`      | Maximum requests allowed per window       |
| `windowSize` | Duration of each window                   |

### Step-by-step
1. Determine current fixed window. Roll forward if needed (previous → current).
2. Calculate weighted estimate using the formula above.
3. If `estimate < limit` → allow request, increment `currCount`.
4. If `estimate >= limit` → reject request.

---

## Benefits Over Other Algorithms

### vs Token Bucket
| Aspect | Sliding Window Counter | Token Bucket |
|--------|----------------------|--------------|
| **Window-based** | ✅ Natural "X per minute" semantics | ❌ Rate-based, harder to reason about |
| **No burst loophole** | ✅ Smoothed estimation prevents burst | ❌ Burst can spike past per-window rate |
| **Distributed** | ✅ Two counters — easy to sync | ⚠️ Float state harder to distribute |

**Key advantage:** **Natural window semantics** — "100 requests per minute" maps directly to the algorithm's parameters. Token Bucket's capacity + refill rate requires more thought to translate into a per-window limit, and its burst capability can exceed the intended per-window count.

### vs Leaky Bucket
| Aspect | Sliding Window Counter | Leaky Bucket |
|--------|----------------------|--------------|
| **Burst tolerance** | ✅ Allows natural traffic patterns | ❌ Forces constant rate |
| **Semantics** | ✅ "N requests per window" — intuitive | ⚠️ "N requests per second drain" — less intuitive |
| **Memory** | ✅ O(1) — two counters | ✅ O(1) — one float |

**Key advantage:** **Allows natural bursty traffic** while still preventing boundary spikes. Leaky Bucket removes all burstiness, which can feel unnecessarily restrictive for legitimate users.

### vs Fixed Window Counter
| Aspect | Sliding Window Counter | Fixed Window Counter |
|--------|----------------------|---------------------|
| **Boundary spike** | ✅ Weighted smoothing prevents 2× burst | ❌ Up to 2× at window boundaries |
| **Accuracy** | ✅ Much closer to true sliding window | ❌ Coarse per-window only |
| **Memory** | ✅ O(1) — just one extra counter | ✅ O(1) — single counter |

**Key advantage:** **Eliminates the boundary spike** with just one extra counter. Fixed Window's 2× boundary spike is its biggest flaw — Sliding Window Counter fixes it with minimal overhead. The weighted approach ensures that traffic at the boundary of two windows is properly accounted for.

### vs Sliding Window Log
| Aspect | Sliding Window Counter | Sliding Window Log |
|--------|----------------------|-------------------|
| **Memory** | ✅ O(1) — two integers | ❌ O(N) — one timestamp per request |
| **Speed** | ✅ O(1) constant time | ❌ O(log N) for eviction |
| **Scalability** | ✅ Handles millions of req/s | ❌ Memory grows with traffic |
| **Distributed** | ✅ Two INCR ops in Redis | ❌ Hard to distribute sorted lists |

**Key advantage:** **O(1) memory and O(1) time** with near-sliding-window accuracy. Sliding Window Log gives perfect accuracy but at O(N) memory cost per user. At 100K req/min per user, that's ~800KB per user — Sliding Window Counter uses ~16 bytes regardless.

---

## Accuracy Analysis

The estimation error is bounded and small in practice:

| Scenario | Max Error |
|----------|-----------|
| Uniform traffic | ~0% |
| All traffic at window start | ≤ limit × (1 / window_segments) |
| All traffic at window end | Similar small bound |
| Worst case | Rarely exceeds 1-2% |

The approximation assumes requests in the previous window were **uniformly distributed**. In practice, this is close enough for virtually all rate limiting use cases.

---

## When to Use Sliding Window Counter

- **Best general-purpose choice** — good balance of accuracy, performance, and simplicity
- **High-traffic APIs** — where O(N) memory (Sliding Window Log) is too expensive
- **Distributed systems** — two Redis `INCR` calls with `EXPIRE` — simple and fast
- **When boundary spike matters but perfect accuracy doesn't** — the sweet spot
- **Most production rate limiters** — this is the go-to algorithm for most real systems

## Real-World Usage

- **Cloudflare** — uses sliding window counter for rate limiting
- **Redis-based rate limiters** — most production implementations use this approach
- **API Gateways** — Kong, Envoy use variants of sliding window counter
- **Social media platforms** — tweet/post rate limiting
