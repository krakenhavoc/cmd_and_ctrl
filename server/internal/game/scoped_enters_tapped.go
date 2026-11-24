package game

import "github.com/google/uuid"

// scoped_enters_tapped.go — "until <duration>, creatures your opponents
// control enter tapped" (#6093, Cone of Cold's 20 row), a CR 614.1c
// replacement effect created by a resolving spell (CR 611.2) and held as a
// scoped record, so it outlives the sorcery that made it and ends with its
// duration rather than with the turn.
//
// Authority of the Consuls prints the same replacement as a static
// ability; this is the resolved, timed form. The record's Controller is
// "you": a creature entering under any other player's control enters
// tapped. The controller a permanent enters under is read the way the
// printed form reads it (LookupCardForEffect on the entering card). The
// mod kind is an on-disk identity: never renamed, never reused.

// ModCreaturesEnterTappedForOpponents is "creatures your opponents control
// enter tapped" for as long as the record lasts. "You" is the record's
// Controller; the mod reads nothing else. A replacement, read by the
// replacement gather rather than the layer pass.
const ModCreaturesEnterTappedForOpponents ModKind = "creaturesEnterTappedForOpponents"

// CreaturesEnterTappedForOpponentsForEffect registers the replacement for
// `controller`'s opponents with duration `d`: "Until your next turn,
// creatures your opponents control enter tapped" is
// d = UntilYourNextTurnDuration(controller). Returns false, registering
// nothing, for a player who is not in the game.
//
// Caller must hold g.mu (write) — every caller is a resolving effect.
func (g *Game) CreaturesEnterTappedForOpponentsForEffect(sourceID, controller uuid.UUID, d Duration, label string) bool {
	if g.playerByIDLocked(controller) == nil {
		return false
	}
	return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeGame, controller,
		[]Mod{{Kind: ModCreaturesEnterTappedForOpponents}}, d, label)
}

// creaturesEnterTappedForOpponentsAppliesLocked is the kind's test: a
// creature moving onto the battlefield under the control of a player other
// than the record's controller. Caller must hold g.mu.
func creaturesEnterTappedForOpponentsAppliesLocked(g *Game, e ScopedEffect, ev *ReplacementEvent) bool {
	if ev.Kind != RepEventMove || ev.NewZone != ZoneBattlefield || ev.Canceled || ev.EntersTapped {
		return false
	}
	entering, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || entering.Controller == e.Controller {
		return false
	}
	return entering.IsCreature()
}
