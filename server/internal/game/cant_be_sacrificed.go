package game

import (
	"fmt"

	"github.com/google/uuid"
)

// cant_be_sacrificed.go: "<this permanent> can't be sacrificed" (#3275).
//
// Two forms, one predicate:
//
//   - The STATIC form is the permanent's own ability, or one granted to
//     it: Alexios, Deimos of Kosmos ("Alexios ... can't be sacrificed"),
//     Hithlain Rope, Assault Suit's equipped creature. A layer-6 static
//     sets Characteristic.CantBeSacrificed, so a permanent that loses
//     all abilities may be sacrificed again (CR 613.1f).
//   - The DURATION form is a resolved effect's rule about one object:
//     Slicer, Hired Muscle's "it can't be sacrificed this turn". It is a
//     ModCantBeSacrificed scoped effect pinned to the object (CR 400.7)
//     and swept at cleanup (CR 514.2), the ModCantBeRegenerated shape.
//     It is not an ability, so losing all abilities does not end it.
//
// CR 101.2: "can't" beats "can", so every route that sacrifices obeys
// it, and it is read in three places:
//
//   - COSTS. A permanent that can't be sacrificed can't pay a sacrifice
//     cost (CR 118.3): validateSacrificeCostLocked refuses it (the one
//     check an activated ability, a mana ability and a spell's
//     additional cost share), the alternative-cost walk leaves it out,
//     and the candidate lists the view, the enumerator and the
//     auto-tapper build are narrowed through SacrificeCostCandidatesForEffect
//     or this predicate, so none of them offers what the engine refuses.
//   - CHOICES. An effect's "sacrifice a creature" (an edict, ADR 0175's
//     ordered collector, a "you may sacrifice") reads
//     sacrificeCandidatesLocked, which leaves it out: it is simply not
//     a legal choice, and a player whose only creature can't be
//     sacrificed has nothing to sacrifice.
//   - THE MOVE. Every sacrifice ends in the sacrifice route, and the
//     route refuses it (sacrificeWithAnswerLocked, the batch filter,
//     routeLegNothingToDoLocked), so "sacrifice it at the next end
//     step", "sacrifice all creatures" and the Saga rule leave it where
//     it is and announce nothing.

// ErrCantBeSacrificed is returned when a permanent that can't be
// sacrificed is asked to be (#3275, CR 101.2). A cost that names it is
// refused with ErrIllegalTarget instead, which this wraps, so a client
// that already reads ErrIllegalTarget reads this the same way.
var ErrCantBeSacrificed = fmt.Errorf("%w: that permanent can't be sacrificed", ErrIllegalTarget)

// permanentCantBeSacrificedLocked reports whether the permanent `c`
// can't be sacrificed right now: its effective characteristics say so
// (the static form), or a live ModCantBeSacrificed record covers this
// object (the duration form). False for nil. Reads only.
//
// Caller must hold g.mu.
func (g *Game) permanentCantBeSacrificedLocked(c *Card) bool {
	if c == nil {
		return false
	}
	if c.Effective().CantBeSacrificed {
		return true
	}
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		if e.Scope == ScopeNone && scopedEffectHasMod(e, ModCantBeSacrificed) && affectedPredicate(e.Affected)(c, g, nil) {
			return true
		}
	}
	return false
}

// cantBeSacrificedLocked is permanentCantBeSacrificedLocked by
// instance ID. False for a card that is not on the battlefield.
//
// Caller must hold g.mu.
func (g *Game) cantBeSacrificedLocked(cardID uuid.UUID) bool {
	return g.permanentCantBeSacrificedLocked(findBattlefieldCard(g, cardID))
}

// CantBeSacrificedForEffect is the exported reader, for card files,
// the view and the legal-move enumerator. False for a card that is not
// on the battlefield.
//
// Caller must hold g.mu (read or write).
func (g *Game) CantBeSacrificedForEffect(cardID uuid.UUID) bool {
	return g.cantBeSacrificedLocked(cardID)
}

// sacrificeableLocked drops every permanent in ids that can't be
// sacrificed. It returns ids itself when nothing is dropped.
//
// Caller must hold g.mu.
func (g *Game) sacrificeableLocked(ids []uuid.UUID) []uuid.UUID {
	for i, id := range ids {
		if !g.cantBeSacrificedLocked(id) {
			continue
		}
		out := append(make([]uuid.UUID, 0, len(ids)-1), ids[:i]...)
		for _, rest := range ids[i+1:] {
			if !g.cantBeSacrificedLocked(rest) {
				out = append(out, rest)
			}
		}
		return out
	}
	return ids
}

// SacrificeCostCandidatesForEffect is SpecCandidatesForEffect's card
// list narrowed to what can pay a sacrifice COST for `playerID`: the
// permanents they control (CR 701.21a) that can be sacrificed (#3275).
// It is the walk every sacrifice-cost picker shares: the view's
// options, the enumerator's pool and the additional-cost payability
// check. It does not target, so hexproof and protection don't matter.
//
// `sourceID` is the permanent whose ability the cost is, whose text the
// clause's words are read through (costSourceLocked, #3872): uuid.Nil
// for a spell's additional cost, which reads its printed words.
//
// Caller must hold g.mu (read or write).
func (g *Game) SacrificeCostCandidatesForEffect(playerID, sourceID uuid.UUID, spec *TargetSpec) []uuid.UUID {
	var out []uuid.UUID
	for _, id := range g.costCandidatesLocked(playerID, sourceID, spec).Cards {
		c := findBattlefieldCard(g, id)
		if c == nil || c.Controller != playerID || g.permanentCantBeSacrificedLocked(c) {
			continue
		}
		out = append(out, id)
	}
	return out
}

// CantBeSacrificedThisTurnForEffect is "<that permanent> can't be
// sacrificed this turn" (Slicer, Hired Muscle; #3275): a
// ModCantBeSacrificed record pinned to the object (CR 400.7), swept at
// cleanup (CR 514.2). A permanent that leaves and returns is a new
// object and may be sacrificed. Registers nothing for a card that is
// not on the battlefield and returns false for it.
//
// Caller must hold g.mu (write).
func (g *Game) CantBeSacrificedThisTurnForEffect(sourceID, cardID uuid.UUID, label string) bool {
	affected := g.PinnedObjectsLocked(cardID)
	if len(affected) == 0 {
		return false
	}
	return g.appendScopedEffectLocked(sourceID, affected, ScopeNone, uuid.Nil,
		[]Mod{{Kind: ModCantBeSacrificed}}, g.PinnedTo(g.UntilEndOfTurnDuration(), cardID),
		label, timeNowUnixNano())
}
