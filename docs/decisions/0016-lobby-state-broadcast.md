# ADR 0016 — Broadcasting lobby mutations to connected clients

**Status:** Implemented · 2026-07-08 · Branch `refactor/major-wip`

## Context

Until now the lobby's HTTP mutations and the in-game WebSocket stream
were two disconnected worlds. `POST /games/{id}/join`,
`/games/{id}/decks`, and `/games/{id}/start` mutated lobby state over
HTTP, but the only way a client learned about the change was to
re-fetch (`GET /games/{id}`) or reload. A player already sitting on
the game page — the common case once a pod is assembling — saw a stale
seat list until they refreshed: a new seat, an opponent's uploaded
deck, or the game flipping to "started" all landed silently.

The in-game action path already solved this: WS actions mutate `Game`
under the write lock and the hub fans a state delta out to every
connected seat, each filtered through the per-connection visibility
rule. We wanted lobby-side mutations to reuse that same fan-out
instead of inventing a second notification channel or making clients
poll.

Two constraints shaped the design:

1. **No import cycle.** `lobby` must not import `ws` — `ws` already
   depends on `lobby` (`lobby.WSAuthorizer`, room lookup). The
   notification seam has to point the other way.
2. **A failed lobby mutation must not broadcast a half-applied
   state.** Lobby mutations touch `Game` state (seating a player,
   installing a deck, starting the game). If a mutation errors partway
   we must not push the partially-mutated view to clients.

## Decisions

### 1. A narrow `StateBroadcaster` interface owned by `lobby`

`lobby` declares the interface it needs and `*ws.Hub` satisfies it:

```go
// lobby package
type StateBroadcaster interface {
    BroadcastState(gameID uuid.UUID, seq uint64, view protocol.GameView)
}
```

`main.go` wires the concrete hub in after both are constructed
(`l.SetStateBroadcaster(hub)`). Dependency direction stays
`ws → lobby`; the lobby only ever sees the interface. The interface is
deliberately one method wide so lobby tests inject a `recordingBroadcaster`
fake and assert on broadcasts without standing up a hub, a manager, and
sockets.

### 2. `Lobby.applyLocked` is the single choke point

Every lobby mutation that changes game state runs through one helper
that, on success, captures the view + seq and calls the broadcaster;
on error it returns without broadcasting. New mutation paths get live
updates for free by routing through it, and the "mutate then notify"
ordering lives in exactly one place instead of being copy-pasted into
each handler.

### 3. `Room.ApplyExternal` applies against a clone and rolls back on error

The lobby doesn't own the game's write lock — the room does. `Room.ApplyExternal(fn)`
is the lobby-side sibling of the WS `Apply` path: it takes the write
lock, snapshots via `Game.Clone()`, runs the mutation, and on error
calls `RestoreFrom(snapshot)` so a failed mutation leaves zero
residue. On success it bumps the room's monotonic `seq` and returns
the filtered `GameView` for the hub to fan out. This is why
`Game.Clone`/`RestoreFrom` were extended (see the deep-clone commit) to
deep-copy turn-scoped replacement effects and per-card `knownBy` sets
and to force a layer recompute on restore — a shallow snapshot would
leak hidden-information knowledge or stale layer state across a
rollback.

### 4. `Game.CurrentState()` for race-free out-of-package gating

Lobby join/start/replay pre-checks previously read `Game.State`
directly. WS actions mutate `State` under the write lock (a concede
flips it to `StateEnded`), so an unlocked read from the lobby
goroutine races. `CurrentState()` takes the read lock and returns the
lifecycle state; out-of-package callers use it instead of touching the
field.

## Consequences

- A player on the game page sees seats fill, decks upload, and the
  game start live, with no reload — the seq watermark on the client
  (see the client auto-reconnect commit) dedupes the broadcast against
  any snapshot it already applied.
- Broadcasts reuse the existing per-connection visibility filter, so
  the lobby fan-out can't leak hidden information any more than the
  in-game path can.
- The wire protocol is unchanged: lobby broadcasts ride the existing
  snapshot/state frame and `seq`. Nothing in `docs/protocol.md`'s v0
  schema changes.
- `Clone`/`RestoreFrom` are now on the hot path for every external
  mutation, not just undo. They must stay a true deep copy; the
  `clone_test.go` isolation tests guard this.

## Alternatives considered

- **Client polling (`GET /games/{id}` on a timer).** Simple, but wasteful
  and laggy, and it duplicates the state the WS stream already carries.
- **A second lobby-events WS channel.** More moving parts and a second
  visibility-filter implementation to keep in sync with the game one.
- **`lobby` imports `ws` directly.** Creates an import cycle and welds
  the two packages together; the interface seam keeps them decoupled
  and testable.
