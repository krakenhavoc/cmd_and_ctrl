# ADR 0001 — WebSocket library for the Go game server

- **Status:** accepted
- **Date:** 2026-04-10
- **Sprint:** S01 — Go server + client scaffold
- **Deciders:** project owner

## Context

The S01 scaffold needs a WebSocket library for the Go game server. PLAN.md §4
specifies "Go (stdlib + [gorilla/websocket](https://github.com/gorilla/websocket)
or similar)". The open question in PLAN.md §7.1 is whether to go with
`gorilla/websocket` or something higher-level like `nhooyr.io/websocket`
(now `coder/websocket`).

The server's needs for S12 are modest:
- Accept N client connections per game (target: 4 players, ≤8 total)
- Receive JSON action frames
- Broadcast JSON state-delta frames
- Graceful close and reconnection
- Ping/pong heartbeats

## Options

### A. `github.com/gorilla/websocket`

The boring, battle-tested choice. Has been the de-facto Go WebSocket library
for a decade. Used by Kubernetes, Caddy, and most of the Go ecosystem.
Stewardship moved to the Gorilla toolkit revival in 2023 after a brief
archive, and is actively maintained again.

**Pros:** maximum ecosystem support, every StackOverflow answer is written
for it, examples are everywhere, known performance characteristics.

**Cons:** lower-level API (manual read/write loops, manual ping/pong timers,
explicit concurrency discipline). Easy to footgun on concurrent writes.

### B. `github.com/coder/websocket` (formerly `nhooyr.io/websocket`)

A newer, more ergonomic library. Uses `context.Context` throughout, has a
cleaner close API, integrates with `net/http` idiomatically. Fewer examples
in the wild but the API is genuinely nicer.

**Pros:** modern Go idioms, context-aware, simpler to use correctly on the
first try.

**Cons:** smaller ecosystem, fewer battle-tested patterns to copy, one less
boring dependency.

### C. `golang.org/x/net/websocket`

Don't. The stdlib adjunct package is explicitly not recommended even in its
own docs, which point at the two options above.

## Decision

**Go with `github.com/gorilla/websocket`.**

Rationale:
- "Boring, battle-tested" is the right posture for a hobby project at
  ~10 hrs/week. Time spent debugging a clever library is time not spent on
  the UX polish that is actually the point.
- Every WebSocket pattern we'll need (hub + client goroutine, pub/sub
  broadcast, reconnect-with-backoff, heartbeat) has a canonical Gorilla
  example. The gorilla/websocket `examples/chat/` directory is within ~50
  lines of what we need for S03.
- The "footgun" concern (concurrent writes) is real but manageable with the
  standard hub pattern — one write goroutine per client, channels for
  incoming messages. We'll encode this in `internal/ws/` from day one.
- If `coder/websocket` proves compelling later, swapping is a few hundred
  lines and not a protocol break.

## Consequences

- Adds `github.com/gorilla/websocket` to `server/go.mod`.
- `internal/ws` will hand-roll the hub pattern (one goroutine per client
  reader, one per writer, central hub for broadcast). Documented inline.
- Revisit if concurrent-write bugs become a pattern, or if context-aware
  cancellation gets awkward.

## References

- [gorilla/websocket](https://github.com/gorilla/websocket)
- [coder/websocket](https://github.com/coder/websocket)
- [gorilla/websocket chat example](https://github.com/gorilla/websocket/tree/main/examples/chat) — the template we'll crib from
