package game

import "github.com/google/uuid"

// targeting_bar.go — "<that creature> can't be the target of spells or
// abilities your opponents control this turn" (Vines of Vastwood, #3860),
// the resolving-spell form of a clause that is on the card's text as a
// STATIC for Tomik (targeting_restriction.go) and as a keyword for
// hexproof.
//
// IT IS NOT HEXPROOF. Hexproof (CR 702.11b) refuses spells and abilities
// the CREATURE'S CONTROLLER's opponents control. Vines of Vastwood's
// clause refuses what the Vines CASTER's opponents control, and the two
// come apart exactly where it matters: Vines cast on a creature
// somebody else controls (a sacrifice-bait line, a blocker an opponent
// is about to lose) locks the creature's own controller out of
// targeting it, and hexproof would not. The first cut of the card
// narrowed the target to "a creature you control" so that hexproof was
// the printed effect word for word, and shipped weaker than printed.
//
// So it is its own record: ModCantBeTargetedByOpponents, an ADR 0041 data
// record over the creature the spell named (locked now, CR 611.2c, and
// following the object not the name: a creature that blinks is a new
// object, CR 400.7), ending at cleanup with the rest of the registry. It
// is read LIVE by the targeting choke point, on every question, so
// nothing is written anywhere and undo, the snapshot and a restore point
// carry it as ordinary registry data. It sits beside the hexproof test
// in canBeTargetedBy and ahead of it, so a "may be targeted as though it
// didn't have hexproof" waiver (Nowhere to Run, Detection Tower) -- which
// waives hexproof and nothing else (hexproof_bypass.go) -- cannot lift it.
//
// Both halves of targeting legality go through canBeTargetedBy: the CR
// 601.2c announce gate and the CR 608.2b resolution re-check, so a spell
// already aimed at the creature when the Vines resolve fizzles later --
// the card's own ruling. "Spells and abilities your opponents control"
// means a spell or ability controlled by anyone but the record's
// Controller, so the Vines' caster (and a creature's controller who is the
// caster) can still target it.

// CantBeTargetedByOpponentsMod is the mod for the record: register it
// over the creature with RegisterScopedEffectForEffect (effects.
// CantBeTargetedByYourOpponentsUntilEOT is the card-facing primitive).
// "You" is the controller of the resolving spell or ability, read off
// the record's source.
func CantBeTargetedByOpponentsMod() Mod { return Mod{Kind: ModCantBeTargetedByOpponents} }

// scopedTargetingBarLocked reports whether a record bars a spell or
// ability controlled by `by` from targeting the permanent `card`: some
// record carries ModCantBeTargetedByOpponents over it, and `by` is not
// that record's Controller.
//
// Nil-receiver safe: the exported CanBeTargetedBy has no game and asks
// through a nil one, which bars nothing. Caller must hold g.mu (read or
// write). Reads only.
func (g *Game) scopedTargetingBarLocked(card *Card, by uuid.UUID) bool {
	if g == nil || card == nil || len(g.ScopedEffects) == 0 {
		return false
	}
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		if e.Controller == by || !scopedEffectHasMod(e, ModCantBeTargetedByOpponents) {
			continue
		}
		if affectedPredicate(e.Affected)(card, g, nil) {
			return true
		}
	}
	return false
}
