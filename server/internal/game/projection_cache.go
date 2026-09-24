package game

import "sync"

// projection_cache.go gives a derived projection of the event log a
// place to live between views (#1401, ADR 0033 amendment 2026-09-24).
//
// The public log (protocol.publicLogOf) is a forward fold over
// Game.Events: the running turn and step, the pending-sacrifice
// marker and the reveal / roll grouping all depend on every event
// before the current one. Re-folding from event 0 on every view made
// each view O(len(Events)) and a game O(events²). The fold's state
// after event N is a pure function of Events[:N], so it can be kept
// and resumed — as long as nobody changes Events[:N] under it.
//
// Two facts make that safe, and both are owned HERE, by the package
// that writes Events:
//
//  1. EmitEvent only ever appends. Nothing writes Events[i] in place.
//  2. The one path that replaces the log with a different history —
//     RestoreFrom, the undo path, which truncates to a snapshot's
//     length and lets the next emit regrow it with the SAME Seq values
//     (eventSeq rewinds) — bumps eventLogGen. A cache that recorded
//     the generation it folded under can therefore tell "the log grew"
//     from "the log was rewound and regrew to the same length".
//
// A game restored from a persisted snapshot, and every Clone, is a
// NEW *Game with a zero cache, so neither needs a bump.
// TestEventLogIsReplacedOnlyWhereTheGenerationMoves holds (1) and (2)
// against the source: an assignment to a Game's Events anywhere else
// fails the test until it bumps the generation or is shown to build
// a fresh game.

// ProjectionCache is an opaque, mutex-guarded slot for one derived
// projection of a game. The game never reads it; the protocol package
// owns what goes in it and how it is validated.
//
// It has its own lock rather than riding g.mu because the projection
// runs under the READ lock (ViewOfGame's ReadSnapshot), and several
// readers — the room's broadcast view and each bot seat's own — may
// build views concurrently. The slot serialises them against each
// other; the read lock they already hold keeps Events still.
type ProjectionCache struct {
	mu sync.Mutex
	v  any
}

// Do runs fn with exclusive access to the slot's value. fn may read,
// replace or clear *slot.
func (c *ProjectionCache) Do(fn func(slot *any)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	fn(&c.v)
}

// LogProjectionCache returns the slot the public-log projection keeps
// its fold state in. The pointer is stable for the life of this *Game
// (RestoreFrom does not replace it), and a Clone gets its own, empty.
func (g *Game) LogProjectionCache() *ProjectionCache {
	return &g.logProjection
}

// EventLogGeneration identifies the HISTORY Events holds, as opposed to
// its length. It changes whenever Events is replaced by a list that is
// not an append-only extension of the previous one, which today means
// RestoreFrom. Two reads that return the same generation, with the
// second log at least as long as the first, saw the second log extend
// the first.
//
// Caller must hold g.mu (read or write).
func (g *Game) EventLogGeneration() uint64 {
	return g.eventLogGen
}
