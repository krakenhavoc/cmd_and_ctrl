package game

import "github.com/google/uuid"

// scoped_statics.go is the registry of continuous effects whose
// lifetime is a DURATION rather than a permanent's presence on the
// battlefield (CR 611.2). It began in S32 as the "until end of turn"
// registry and grew the other three CR 611.2 durations in S38
// (ADR 0063, #755) — hence the rename from turn_scoped_statics.go: a
// registry that can hold an Agent of Treachery for the rest of the
// game is not "turn-scoped".
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
// `Game`, appended to under the resolution write lock, swept at
// known moments, carried by `clone.go`. A parallel registry rather
// than a reuse of that one because the two feed different pipelines:
// a replacement is consulted per *event* by
// `gatherActiveReplacementsLocked`, a static is consulted per
// *recompute pass* by `activeStaticAbilitiesLocked`, and the two
// have no field in common beyond the duration.
//
// What is NOT here:
//   - Layer 1 (copy) with a duration (Mirage Mirror, Cytoshape).
//     The bucket exists (ADR 0043 §5) and nothing registers into it.
//   - Durations for the REPLACEMENT twin. `TurnScopedReplacements`
//     is still cleared wholesale at cleanup and has no duration
//     field; no card needs a longer-lived replacement yet
//     (ADR 0063 Decision 8).

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
// a clone would share the pointer with the original. `Duration` is
// held to the same standard and is stricter still: it is plain data,
// with no closure at all (duration.go).
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
	//
	// Two effects created by ONE spell share a timestamp on purpose:
	// that is what makes an exchange of control a single effect for
	// CR 613.7 (CR 701.12, `ExchangeControlForEffect`).
	Timestamp int64

	// Duration is how long the effect lasts (CR 611.2). Plain data:
	// see duration.go for the four kinds and for
	// `durationExpiredLocked`, the one function that decides when
	// any continuous effect in the game is over.
	Duration Duration

	// Label is human-readable attribution for logs and tests
	// ("Giant Growth — +3/+3"). Not on the wire.
	Label string
}

// RegisterScopedStaticForEffect installs `ability` as a floating
// continuous effect with the given duration, and bumps the layer
// version so the next recompute picks it up.
//
// `sourceID` names the card that created the effect — the resolving
// spell for Giant Growth, the source permanent for a loyalty or
// activated ability. It is looked up across every tracked zone and
// stored as a value copy; a lookup miss stores a stub carrying just
// the instance ID, which is harmless for an ability whose
// predicates don't read `source` (the pinned-target case) and is
// the caller's problem for one that does.
//
// Build the duration with one of the constructors in duration.go —
// `g.UntilEndOfTurnDuration()`, `g.UntilYourNextTurnDuration(p)`,
// `g.ForAsLongAsOnBattlefieldDuration(src)`, `IndefiniteDuration()`.
// The zero value is "until end of turn", so a caller that forgets
// gets the safest answer rather than an effect that never ends.
//
// CREATED DURING THE END STEP. "Until end of turn" means this
// turn's cleanup step even when the effect is created in the end
// step — the end step is not the end of the turn, cleanup is
// (CR 514.2). The cleanup sweep drops every `UntilEndOfTurn` entry
// it sees, so a grant made in the end step dies a few moments later
// at that same turn's cleanup. A duration keyed on "the next
// cleanup I see after the turn I was made in" would instead leak
// the grant into the following turn, which is the bug ADR 0035 §3
// exists to prevent.
//
// Caller must hold g.mu (write). Effects call this from inside the
// resolution frame, which already holds it.
func (g *Game) RegisterScopedStaticForEffect(ability StaticAbility, sourceID uuid.UUID, label string, d Duration) {
	g.registerScopedStaticLocked(ability, sourceID, label, d, timeNowUnixNano())
}

// registerScopedStaticLocked is the shared body, with the CR 613.7
// timestamp passed in. Unexported because there is exactly one
// reason to choose a timestamp rather than read the clock —
// registering the two halves of an exchange as ONE effect
// (CR 701.12) — and that caller lives in this package.
//
// Caller must hold g.mu (write).
func (g *Game) registerScopedStaticLocked(ability StaticAbility, sourceID uuid.UUID, label string, d Duration, ts int64) {
	src := Card{InstanceID: sourceID}
	if found, ok := g.LookupCardForEffect(sourceID); ok {
		src = cloneCard(found)
	}
	g.ScopedStatics = append(g.ScopedStatics, ScopedStatic{
		Ability:   ability,
		Source:    src,
		Timestamp: ts,
		Duration:  d,
		Label:     label,
	})
	g.layerVersion.Add(1)
}

// ClearEndOfTurnScopedStaticsLocked is the CR 514.2 cleanup sweep:
// "until end of turn" effects end during the cleanup step. Called
// from `sweepTurnEndLocked` alongside the damage wipe, the
// turn-scoped replacement clear and the impulse-exile sweep — the
// same "this turn is over" pass. It drops effects of every other
// duration whose time has also run out, because it is the same
// sweep.
//
// Caller must hold g.mu.
func (g *Game) ClearEndOfTurnScopedStaticsLocked() {
	g.sweepScopedStaticsLocked(true)
	// #1197: granted PLAYER abilities ride the same schedule. They
	// are not continuous effects over objects and so are not in the
	// registry above, but they carry the same Duration and are ended
	// by the same durationExpiredLocked, and a second schedule would
	// be a second thing to keep in step. See player_statics.go.
	g.sweepPlayerStaticsLocked(true)
}

// ClearExpiredScopedStaticsLocked is the ordinary sweep: it drops
// every scoped static whose duration has run out, WITHOUT treating
// the current moment as the end of a turn. Called when a turn begins
// (the "until your next turn" boundary) and at the top of every
// layer recompute (where a "for as long as" condition can have gone
// false since the last pass).
//
// Caller must hold g.mu.
func (g *Game) ClearExpiredScopedStaticsLocked() {
	g.sweepScopedStaticsLocked(false)
	// #1197, as above: the "until your next turn" boundary is exactly
	// where Teferi's Protection and The One Ring's shield end.
	g.sweepPlayerStaticsLocked(false)
}

// sweepScopedStaticsLocked drops expired entries and bumps the layer
// version when it drops any, so the next recompute rebuilds without
// them. `endOfTurn` is handed straight to `durationExpiredLocked`,
// which is the only code that decides what a duration means.
//
// Allocates a fresh slice rather than filtering in place with
// `s[:0]`: the backing array is shared with every undo snapshot
// `Clone` has taken, so an in-place compaction would rewrite
// history. That is the same trap `cloneCard` exists to avoid on the
// card side.
//
// The sweep is idempotent, which matters because CR 514.3a can give a
// turn a second cleanup step and the hook runs again for it (#661),
// because the cleanup hook runs again after a discard pause drains,
// and because the layer recompute runs it on every pass.
//
// Caller must hold g.mu.
func (g *Game) sweepScopedStaticsLocked(endOfTurn bool) {
	// ADR 0041 phase 3's data records ride the same sweep (#1497).
	if g.sweepScopedEffectsLocked(endOfTurn) {
		g.layerVersion.Add(1)
	}
	if len(g.ScopedStatics) == 0 {
		return
	}
	kept := make([]ScopedStatic, 0, len(g.ScopedStatics))
	for _, s := range g.ScopedStatics {
		if !g.durationExpiredLocked(s.Duration, endOfTurn) {
			kept = append(kept, s)
		}
	}
	if len(kept) == len(g.ScopedStatics) {
		return
	}
	if len(kept) == 0 {
		kept = nil
	}
	g.ScopedStatics = kept
	g.layerVersion.Add(1)
}

// scopedContinuousEffectsLocked adapts the registry into the
// `ContinuousEffect` values the layer engine sorts and applies.
// Reuses `staticContinuousEffect` — a scoped static differs from a
// battlefield static only in where its source and timestamp come
// from, so it needs no second adapter and lands in the same
// timestamp sort.
//
// The `source` pointers reference into `g.ScopedStatics`, with
// the same lifetime contract as the battlefield pointers
// `activeStaticAbilitiesLocked` hands out: valid for the duration
// of one recompute pass, never retained across a mutation.
//
// Caller must hold g.mu in write mode (the recompute pass does).
func (g *Game) scopedContinuousEffectsLocked() []ContinuousEffect {
	// The data records first (#1497); order within a bucket is decided
	// by timestamp, not by which registry an effect came from.
	data := g.scopedEffectContinuousEffectsLocked()
	if len(g.ScopedStatics) == 0 {
		return data
	}
	out := make([]ContinuousEffect, 0, len(data)+len(g.ScopedStatics))
	out = append(out, data...)
	for i := range g.ScopedStatics {
		s := &g.ScopedStatics[i]
		out = append(out, staticContinuousEffect{
			ability:   s.Ability,
			source:    &s.Source,
			timestamp: s.Timestamp,
		})
	}
	return out
}
