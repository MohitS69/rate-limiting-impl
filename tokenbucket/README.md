# Token Bucket Algorithm

## How It Works

The Token Bucket maintains a "bucket" that holds **tokens**. Tokens are added at a constant **refill rate** (e.g., 10 tokens/sec). Each incoming request must consume one or more tokens. If enough tokens are available, the request is **allowed**; otherwise it is **rejected**.

```
┌────────────────────────────────────────┐
│          TOKEN BUCKET                  │
│                                        │
│   Refill Rate ──►  [●●●●●○○○○○]       │
│   (constant)        ▲          │       │
│                     │    capacity=10   │
│                     │                  │
│              Request arrives           │
│              ─ Has tokens? ──► ALLOW   │
│              ─ No tokens?  ──► REJECT  │
└────────────────────────────────────────┘
```

### Key Parameters
| Parameter    | Description                                |
|--------------|--------------------------------------------|
| `capacity`   | Maximum tokens the bucket can hold (burst)  |
| `refillRate` | Tokens added per second                     |

### Step-by-step
1. Bucket starts full (capacity tokens).
2. Each request consumes 1 token (or N tokens for variable-cost requests).
3. Tokens are replenished at `refillRate` tokens/second.
4. Tokens never exceed `capacity`.
5. If `tokens < cost`, request is rejected.

---

## Benefits Over Other Algorithms

### vs Leaky Bucket
| Aspect | Token Bucket | Leaky Bucket |
|--------|-------------|--------------|
| **Burst traffic** | ✅ Allows controlled bursts up to capacity | ❌ Smooths everything to a constant rate |
| **Flexibility** | ✅ Variable-cost requests (`AllowN`) | ❌ Each request is equal |
| **Use case** | APIs that tolerate short bursts | Scenarios requiring constant throughput |

**Key advantage:** Token Bucket **permits bursts** — ideal for real-world API traffic that is naturally bursty. A user might legitimately need to make 5 rapid requests, and that's fine as long as the average rate stays within limits.

### vs Fixed Window Counter
| Aspect | Token Bucket | Fixed Window Counter |
|--------|-------------|---------------------|
| **Boundary spike** | ✅ No boundary problem | ❌ 2× burst at window edges |
| **Granularity** | ✅ Per-request decision | ❌ Coarse per-window counter |
| **Memory** | ✅ O(1) — two floats + timestamp | ✅ O(1) — one counter |

**Key advantage:** No **boundary spike** problem. Fixed Window can allow 2× the limit if requests cluster at the edge of two adjacent windows.

### vs Sliding Window Log
| Aspect | Token Bucket | Sliding Window Log |
|--------|-------------|-------------------|
| **Memory** | ✅ O(1) | ❌ O(N) — stores every timestamp |
| **Accuracy** | ⚠️ Rate-based, not count-based | ✅ Pixel-perfect accuracy |
| **Burst support** | ✅ Built-in via capacity | ❌ No burst concept |

**Key advantage:** **Constant memory** regardless of traffic volume. Sliding Window Log stores one timestamp per request — at high QPS, memory can explode.

### vs Sliding Window Counter
| Aspect | Token Bucket | Sliding Window Counter |
|--------|-------------|----------------------|
| **Burst** | ✅ Explicit burst via capacity | ❌ No burst support |
| **Simplicity** | ✅ Intuitive model | ⚠️ Weighted estimation can confuse |
| **Accuracy** | ⚠️ Rate-based | ⚠️ Approximation |

**Key advantage:** **Explicit, tunable burst** control. You set exactly how bursty traffic can be via capacity, rather than relying on statistical estimation.

---

## When to Use Token Bucket

- **APIs with bursty traffic patterns** — users make a few rapid calls, then go idle
- **Variable-cost operations** — different endpoints cost different amounts
- **Rate shaping at network level** — widely used in networking (Linux `tc`, AWS API Gateway)
- **When you need both average rate AND burst control** in a single knob

## Real-World Usage

- **AWS API Gateway** — default rate limiter
- **Google Cloud** — many services use token bucket internally
- **Linux Traffic Control (`tc`)** — token bucket filter (TBF)
- **Stripe, GitHub API** — token-bucket-style rate limiting
