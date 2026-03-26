# Sliding Window Log Algorithm

## How It Works

The Sliding Window Log maintains a **sorted list of timestamps** — one for each request. When a new request arrives:
1. All timestamps older than `now - windowSize` are **evicted**.
2. If the remaining count is under the limit, the request is **allowed** and its timestamp is appended.
3. Otherwise, the request is **rejected**.

```
  Window = 60 seconds
  now = 12:01:30

  Log (after eviction):
  ┌──────────────────────────────────────────────────┐
  │ 12:00:31  12:00:45  12:01:02  12:01:15  12:01:28 │
  │                                                    │
  │ count = 5,  limit = 10  →  ALLOW (5 < 10)         │
  │ → append 12:01:30 to log                           │
  └──────────────────────────────────────────────────┘

  Evicted (before 12:00:30):
  ┌─────────────────────────┐
  │ 12:00:05  12:00:12  ... │  ← removed
  └─────────────────────────┘
```

### Key Parameters
| Parameter    | Description                                    |
|--------------|------------------------------------------------|
| `limit`      | Maximum requests allowed within the window      |
| `windowSize` | Duration of the sliding window (e.g., 1 minute) |

### Step-by-step
1. Record current timestamp.
2. Evict all log entries older than `now - windowSize`.
3. Count remaining entries.
4. If `count < limit` → allow and append timestamp.
5. If `count >= limit` → reject.

---

## Benefits Over Other Algorithms

### vs Token Bucket
| Aspect | Sliding Window Log | Token Bucket |
|--------|-------------------|--------------|
| **Accuracy** | ✅ Exact request count per window | ❌ Rate-based, not count-based |
| **No burst loophole** | ✅ Strict per-window count | ❌ Burst can exceed per-window average |
| **Auditability** | ✅ Full request history available | ❌ No history |

**Key advantage:** **Pixel-perfect accuracy**. The count within any sliding window is *exact* — not estimated, not rate-approximated. If your limit is 100 requests/minute, there will **never** be more than 100 requests in any 60-second span. Token Bucket's burst mechanism can temporarily exceed the per-window rate.

### vs Leaky Bucket
| Aspect | Sliding Window Log | Leaky Bucket |
|--------|-------------------|--------------|
| **Burst tolerance** | ✅ Allows natural bursts within limit | ❌ Forces constant output rate |
| **Flexibility** | ✅ Only limits count, not pattern | ❌ Shapes output to constant rate |
| **Audit trail** | ✅ Full timestamp log | ❌ No history |

**Key advantage:** **Permits natural traffic patterns** while guaranteeing strict limits. Leaky Bucket forces an artificial constant rate — Sliding Window Log lets traffic be bursty as long as the total count stays within the window limit.

### vs Fixed Window Counter
| Aspect | Sliding Window Log | Fixed Window Counter |
|--------|-------------------|---------------------|
| **Boundary spike** | ✅ No boundary problem — window slides | ❌ 2× burst at window edges |
| **Accuracy** | ✅ Exact count in any window span | ❌ Only accurate within fixed windows |
| **Fairness** | ✅ Consistent behavior regardless of timing | ❌ Lucky users at boundaries get 2× |

**Key advantage:** **Eliminates the boundary spike problem completely**. Fixed Window's most critical flaw — allowing 2× traffic at window boundaries — simply cannot happen with a true sliding window that tracks every timestamp.

### vs Sliding Window Counter
| Aspect | Sliding Window Log | Sliding Window Counter |
|--------|-------------------|----------------------|
| **Accuracy** | ✅ Exact count — no estimation | ❌ Weighted approximation |
| **Strictness** | ✅ Guaranteed never exceeds limit | ⚠️ Can slightly exceed limit |
| **Audit trail** | ✅ Full timestamp history | ❌ Only counters |

**Key advantage:** **Guaranteed correctness**. Sliding Window Counter uses weighted interpolation that can occasionally let through slightly more than the limit. Sliding Window Log is the only algorithm that provides a **mathematical guarantee** that the limit is never exceeded in any window.

---

## Tradeoffs

The Sliding Window Log's main cost is **memory**:

| Traffic Rate | Memory per User |
|-------------|----------------|
| 10 req/min  | ~80 bytes      |
| 1K req/min  | ~8 KB          |
| 100K req/min| ~800 KB        |
| 1M req/min  | ~8 MB          |

Each timestamp is ~8 bytes (`time.Time` in Go). At very high request rates, this adds up.

---

## When to Use Sliding Window Log

- **Strict compliance requirements** — financial APIs, regulatory limits where exceeding the limit even briefly is unacceptable
- **Security-critical rate limiting** — brute force protection where accuracy is paramount
- **Low-to-moderate traffic** — where the O(N) memory cost is acceptable
- **Audit trail needed** — when you also need to know *when* each request happened
- **When correctness > performance** — the gold standard for accuracy

## Real-World Usage

- **Payment processors** — strict transaction rate limits
- **Authentication systems** — login attempt limiting where 2× at boundary is a security hole
- **SMS/notification services** — strict per-user send rate limits
- **Compliance-driven APIs** — financial trading rate limits mandated by regulators
