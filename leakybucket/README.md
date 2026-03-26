# Leaky Bucket Algorithm

## How It Works

The Leaky Bucket acts like a **queue with a fixed drain rate**. Incoming requests fill the bucket. The bucket "leaks" (processes requests) at a **constant, steady rate**. If the bucket overflows (reaches capacity), new requests are **rejected**.

```
┌────────────────────────────────────────┐
│          LEAKY BUCKET                  │
│                                        │
│   Requests ──►  [████████░░]           │
│   (bursty)       │    capacity=10      │
│                  │                     │
│                  ▼                     │
│            Constant Leak               │
│            (e.g., 5 req/s)             │
│                  │                     │
│                  ▼                     │
│            Processed Output            │
│            (perfectly smooth)          │
└────────────────────────────────────────┘
```

### Key Parameters
| Parameter   | Description                                 |
|-------------|---------------------------------------------|
| `capacity`  | Maximum pending requests the bucket can hold |
| `leakRate`  | Requests drained (processed) per second      |

### Step-by-step
1. Bucket starts empty.
2. Each incoming request adds 1 unit of "water".
3. Water drains at `leakRate` units/second (constant).
4. If `water + 1 > capacity`, request is **rejected**.
5. Output rate is always constant, regardless of input burstiness.

---

## Benefits Over Other Algorithms

### vs Token Bucket
| Aspect | Leaky Bucket | Token Bucket |
|--------|-------------|--------------|
| **Output rate** | ✅ Perfectly constant & smooth | ❌ Allows bursts |
| **Traffic shaping** | ✅ True shaping — smooths output | ⚠️ Only policing — allows or rejects |
| **Predictability** | ✅ Downstream always sees uniform load | ❌ Downstream can see spikes |

**Key advantage:** The Leaky Bucket produces a **perfectly smooth output rate**. This is critical when downstream services cannot handle any burst — e.g., a database that chokes on sudden spikes. Token Bucket allows bursts, which can overwhelm fragile backends.

### vs Fixed Window Counter
| Aspect | Leaky Bucket | Fixed Window Counter |
|--------|-------------|---------------------|
| **Smoothness** | ✅ Constant drain, no spikes | ❌ All requests can cluster at window start |
| **Boundary spike** | ✅ No boundary problem | ❌ 2× spike at window edges |
| **Queuing** | ✅ Inherently queues excess requests | ❌ No queuing, just accept/reject |

**Key advantage:** Leaky Bucket **eliminates the boundary spike problem** entirely and provides inherent request queuing, while Fixed Window has no concept of smoothing at all.

### vs Sliding Window Log
| Aspect | Leaky Bucket | Sliding Window Log |
|--------|-------------|-------------------|
| **Memory** | ✅ O(1) | ❌ O(N) per user |
| **Output pattern** | ✅ Constant rate | ⚠️ Allows clustering within window |
| **Simplicity** | ✅ Simple state machine | ⚠️ Requires sorted timestamp list |

**Key advantage:** **Constant memory** and **constant output** — Sliding Window Log only limits the *count* within a window but doesn't smooth the output pattern.

### vs Sliding Window Counter
| Aspect | Leaky Bucket | Sliding Window Counter |
|--------|-------------|----------------------|
| **Output smoothing** | ✅ True constant rate | ❌ No output smoothing |
| **Deterministic** | ✅ Exact behavior, no estimation | ⚠️ Weighted approximation |
| **Memory** | ✅ O(1) | ✅ O(1) |

**Key advantage:** **Deterministic behavior** — no weighted approximation. The output rate is exactly `leakRate` requests/second, always.

---

## When to Use Leaky Bucket

- **Protecting fragile backends** — database or downstream service can't handle any burst
- **Network traffic shaping** — producing constant bit rate output
- **Message queue consumers** — processing messages at a steady pace
- **When predictable, uniform throughput** is more important than allowing bursts

## Real-World Usage

- **Network routers** — traffic shaping to enforce constant bit rate
- **ATM networks** — the leaky bucket was originally designed for ATM traffic shaping
- **NGINX** — `limit_req` directive uses a leaky bucket internally
- **Message queue consumers** — processing at steady rate to avoid overwhelming workers
