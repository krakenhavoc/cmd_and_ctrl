package game

import "github.com/google/uuid"

// turn_scoped_statics.go is the S32 "until end of turn" continuous-
// effect registry (CR 611.2 one-shot continuous effects, CR 514.2
// expiry).
//
// Why this file exists: every continuous effect the layer engine
// knew about before S32 was sourced from a permanent on the
// battlefield. `activeStaticAbilitiesLocked` walks `g.Battlefield`
// and looks each card's statics up in the catalog, so a static
// ability exists exactly as long as its source permanent does
// (CR 113.6). That is the right model for an anthem and the wrong
// model for Giant Growth: the spell is in the graveyard a moment
// after it resolves, and its +3/+3 has to outlive it.
//
// The shape here is a deliberate copy of `TurnScopedReplacements`
// (replacements.go), the Fog machinery — a plain exported slice on
// `Game`, appended to under the resolution write lock, swept at the
// cleanup-step entry hook, carried by `clone.go`. A parallel
// registry rather than a reuse of that one because the two feed
// different pipelines: a replacement is consulted per *event* by
// `gatherActiveReplacementsLocked`, a static is consulted per
// *recompute pass* by `activeStaticAbilitiesLocked`, and the two
// have no field in common beyond the duration.
//
// What is NOT here:
//   - Durations other than "until end of turn". "Until your next
//     turn" (Teferi's Protection), "for as long as ~ remains
//     exiled", and "until ~ leaves the battlefield" all want the
//     same registry with a different expiry predicate;
//     `ExpiresAfterTurn` is the seam, with the round-vs-seat-turn
//     caveat documented on that field. No catalog card needs them
//     yet, so no speculative field.
//   - Layer 1 (copy) and layer 2 (control). Both are still deferred
//     engine-wide (ADR 0012), and a turn-scoped duration doesn't
//     change that — a UEOT control-change effect would need layer 2
//     to exist first.

// ScopedStatic is one continuous effect whose lifetime is a
// duration rather than a permanent's presence on the battlefield.
//
// The `Ability` is the same `StaticAbility` a catalog card declares
// in `Spec.Static`, so the layer engine needs no second vocabulary:
// the recompute pass adapts both through `staticContinuousEffect`
// and sorts them together by timestamp within each layer bucket
// (CR 613.7).
//
// IMMUTABILITY CONTRACT. Every field is written once at
// registration and never mutated afterwards. `Clone` relies on
// this: it copies the slice into a fresh backing array but shares
// the `Source` card's inner slices, exactly as the replacement
// registry and the LKI map do. Mutating a stored ScopedStatic in
// place would leak across an undo snapshot.
//
// CLOSURE CONTRACT. `Ability.AppliesTo` / `Ability.Apply` outlive
// the effect's source and may outlive any particular battlefield
// layout, so they must capture only values — instance IDs, player
// IDs, ints. A closure that captures a `*Card` pointer into
// `g.Battlefield.Cards` dangles the moment a permanent moves, and
// a clone would share the pointer with the original.
type ScopedStatic struct {
	// Ability is the continuous effect itself.
	Ability StaticAbility

	// Source is a value copy of the card that created the effect,
	// captured at registration — last known information, because
	// the real card is usually in the graveyard by the time the
	// effect is consulted. Handed to `AppliesTo` / `Apply` as their
	// `source` argument so a "creatures you control" predicate can
	// read `source.Controller` the way a battlefield static does.
	Source Card

	// Timestamp is the effect's CR 613.7 timestamp: the moment it
	// was created, on the same Unix-nano clock the battlefield uses
	// for `EnteredBattlefieldAt`. A pump created after an anthem
	// entered therefore sorts after it within layer 7c, and a
	// second "base P/T becomes N/N" created later overwrites an
	// earlier one within 7b.
	Timestamp int64

	// ExpiresAfterTurn is the turn number this effect survives
	// through. The cleanup-step sweep drops every entry whose
	// ExpiresAfterTurn is not greater than the current turn number.
	// "Until end of turn" stamps the turn number current at
	// registration, which is always <= the number at the next
	// cleanup, so every "until end of turn" grant is dropped by the
	// first cleanup step it sees. That is the point: it is the
	// EARLIEST cleanup that ends the effect, including the one at
	// the end of the very turn it was made in.
	//
	// CAVEAT for anything longer. `Turn.Number` counts ROUNDS, not
	// seat-turns — it only increments when the cursor wraps back to
	// seat 0 (turn.go `advance`), so in a four-player game all four
	// seats share a number. That is invisible to "until end of turn"
	// but it means a future "until your next turn" duration cannot be
	// expressed by bumping this field alone; it would need the
	// active seat alongside it.
	ExpiresAfterTurn int

	// Label is human-readable attribution for logs and tests
	// ("Giant Growth — +3/+3"). Not on the wire.
	Label string
}

// RegisterTurnScopedStaticForEffect installs `ability` as a floating
// continuous effect that expires at the end of the current turn
// (CR 514.2), and bumps the layer version so the next recompute
// picks it up.
//
// `sourceID` names the card that created the effect — the resolving
// spell for Giant Growth, the source permanent for a loyalty or
// activated ability. It is looked up across every tracked zone and
// stored as a value copy; a lookup miss stores a stub carrying just
// the instance ID, which is harmless for an ability whose
// predicates don't read `source` (the pinned-target case) and is
// the caller's problem for one that does.
//
// CREATED DURING THE END STEP. "Until end of turn" means this
// turn's cleanup step even when the effect is created in the end
// step — the end step is not the end of the turn, cleanup is
// (CR 514.2). Stamping `ExpiresAfterTurn` with the turn number
// current at registration gets that right for free: the sweep at
// this same turn's cleanup sees `ExpiresAfterTurn == Turn.Number`
// and drops it. A duration keyed on "the next cleanup I see after
// the turn I was made in" would instead leak the grant into the
// following turn, which is the bug this field exists to prevent.
//
// Caller must hold g.mu (write). Effects call this from inside the
// resolution frame, which already holds it.
func (g *Game) RegisterTurnScopedStaticForEffect(ability StaticAbility, sourceID uuid.UUID, label string) {
	src := Card{InstanceID: sourceID}
	if found, ok := g.LookupCardForEffect(sourceID); ok {
		src = cloneCard(found)
	}
	g.TurnScopedStatics = append(g.TurnScopedStatics, ScopedStatic{
		Ability:          ability,
		Source:           src,
		Timestamp:        timeNowUnixNano(),
		ExpiresAfterTurn: g.Turn.Number,
		Label:            label,
	})
	g.layerVersion.Add(1)
}

// ClearExpiredTurnScopedStaticsLocked drops every scoped static
// whose duration has run out, and bumps the layer version when it
// drops any so the next recompute rebuilds without them. Called
// from the cleanup-step entry hook alongside the damage wipe, the
// turn-scoped replacement clear and the impulse-exile sweep — the
// same "this turn is over" pass.
//
// Allocates a fresh slice rather than filtering in place with
// `s[:0]`: the backing array is shared with every undo snapshot
// `Clone` has taken, so an in-place compaction would rewrite
// history. That is the same trap `cloneCard` exists to avoid on the
// card side.
//
// The sweep is idempotent, which matters because the cleanup hook
// runs again after a discard pause drains.
//
// Caller must hold g.mu.
func (g *Game) ClearExpiredTurnScopedStaticsLocked() {
	if len(g.TurnScopedStatics) == 0 {
		return
	}
	kept := make([]ScopedStatic, 0, len(g.TurnScopedStatics))
	for _, s := range g.TurnScopedStatics {
		if s.ExpiresAfterTurn > g.Turn.Number {
			kept = append(kept, s)
		}
	}
	if len(kept) == len(g.TurnScopedStatics) {
		return
	}
	if len(kept) == 0 {
		kept = nil
	}
	g.TurnScopedStatics = kept
	g.layerVersion.Add(1)
}

// turnScopedContinuousEffectsLocked adapts the registry into the
// `ContinuousEffect` values the layer engine sorts and applies.
// Reuses `staticContinuousEffect` — a scoped static differs from a
// battlefield static only in where its source and timestamp come
// from, so it needs no second adapter and lands in the same
// timestamp sort.
//
// The `source` pointers reference into `g.TurnScopedStatics`, with
// the same lifetime contract as the battlefield pointers
// `activeStaticAbilitiesLocked` hands out: valid for the duration
// of one recompute pass, never retained across a mutation.
//
// Caller must hold g.mu in write mode (the recompute pass does).
func (g *Game) turnScopedContinuousEffectsLocked() []ContinuousEffect {
	if len(g.TurnScopedStatics) == 0 {
		return nil
	}
	out := make([]ContinuousEffect, 0, len(g.TurnScopedStatics))
	for i := range g.TurnScopedStatics {
		s := &g.TurnScopedStatics[i]
		out = append(out, staticContinuousEffect{
			ability:   s.Ability,
			source:    &s.Source,
			timestamp: s.Timestamp,
		})
	}
	return out
}
