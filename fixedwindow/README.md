# Fixed Window Counter Algorithm

## How It Works

The Fixed Window Counter divides time into **aligned, non-overlapping windows** of a fixed size (e.g., every minute from :00 to :59). A counter tracks requests in the current window. When the window expires, the counter resets to zero.

```
  Window 1 (12:00-12:01)    Window 2 (12:01-12:02)
┌────────────────────────┐ ┌────────────────────────┐
│  counter: 8 / limit:10 │ │  counter: 0 / limit:10 │
│  ████████░░             │ │  ░░░░░░░░░░             │
│                         │ │                         │
│  8 requests allowed     │ │  Counter reset!         │
│  2 more available       │ │  10 available           │
└────────────────────────┘ └────────────────────────┘
         time ──────────────────►
```

### Key Parameters
| Parameter    | Description                              |
|--------------|------------------------------------------|
| `limit`      | Maximum requests allowed per window       |
| `windowSize` | Duration of each fixed window             |

### Step-by-step
1. Determine which window the current time falls into.
2. If new window → reset counter to 0.
3. If `counter < limit` → allow request, increment counter.
4. If `counter >= limit` → reject request.

---

## The Boundary Spike Problem

The main weakness of Fixed Window is the **boundary spike**. At the junction of two windows, a user can make `limit` requests at the end of window 1 and `limit` requests at the start of window 2 — effectively getting **2× the limit** in a short burst.

```
Window 1                    Window 2
    ............████████████████████............
                ▲               ▲
          10 requests at   10 requests at
          end of W1        start of W2
          
          = 20 requests in ~1 second span!
            (but limit is 10/minute)
```

---

## Benefits Over Other Algorithms

### vs Token Bucket
| Aspect | Fixed Window | Token Bucket |
|--------|-------------|--------------|
| **Simplicity** | ✅ Dead simple — one counter | ⚠️ Needs float math, refill logic |
| **Implementation** | ✅ Trivial to distribute (Redis INCR + TTL) | ⚠️ Harder in distributed systems |
| **Overhead** | ✅ Minimal CPU — just an increment | ⚠️ Needs time-based refill calculation |

**Key advantage:** **Extreme simplicity and distributed-friendliness**. A Fixed Window Counter can be implemented in Redis with just `INCR` + `EXPIRE` — two atomic operations. Token Bucket requires compare-and-swap on floats.

### vs Leaky Bucket
| Aspect | Fixed Window | Leaky Bucket |
|--------|-------------|--------------|
| **Burst tolerance** | ✅ Allows natural bursts within window | ❌ Forces constant rate |
| **Simplicity** | ✅ Single integer counter | ⚠️ Needs drain calculation |
| **No queuing needed** | ✅ Accept/reject instantly | ⚠️ Conceptually a queue |

**Key advantage:** No output-shaping/queuing overhead. Fixed Window is a **pure accept/reject gate** — the simplest mental model for rate limiting.

### vs Sliding Window Log
| Aspect | Fixed Window | Sliding Window Log |
|--------|-------------|-------------------|
| **Memory** | ✅ O(1) — single counter | ❌ O(N) — one entry per request |
| **Speed** | ✅ O(1) per request | ❌ O(log N) eviction |
| **Distributed** | ✅ Trivial with Redis INCR | ❌ Hard to distribute sorted lists |

**Key advantage:** **O(1) everything** — memory, time, and trivial distributed implementation. Sliding Window Log stores every timestamp, which becomes expensive at scale.

### vs Sliding Window Counter
| Aspect | Fixed Window | Sliding Window Counter |
|--------|-------------|----------------------|
| **Simplicity** | ✅ Single counter, no math | ⚠️ Weighted interpolation logic |
| **Deterministic** | ✅ Exact count, no estimation | ❌ Approximate weighted count |
| **Speed** | ✅ O(1), pure integer ops | ✅ O(1), but with float math |

**Key advantage:** **No approximation** — the count is exact within the window. Sliding Window Counter uses estimation that can occasionally allow slightly more than the limit.

---

## When to Use Fixed Window Counter

- **High-throughput, low-complexity scenarios** — where simplicity matters most
- **Distributed systems** — when you need Redis/Memcached compatibility with minimal operations
- **Coarse-grained limiting** — daily/hourly API quotas where boundary spikes don't matter
- **When you can tolerate the boundary spike** — limits like "1000 requests/hour" where a brief 2× burst is acceptable

## Real-World Usage

- **GitHub API** — hourly rate limits use fixed window counters
- **Many REST APIs** — use fixed windows for daily/hourly quotas
- **CDN rate limiting** — per-minute request caps
- **Redis-based rate limiters** — `INCR` + `EXPIRE` is the canonical implementation
