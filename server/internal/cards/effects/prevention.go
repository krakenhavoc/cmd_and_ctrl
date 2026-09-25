package effects

import (
	"github.com/google/uuid"
)

// prevention.go — CR 615 damage prevention, the two shapes it comes
// in and the one piece of state that separates them.
//
// S17 shipped Fog as a proof of concept: a shield that cancels every
// combat-damage event for the rest of the turn. That is a SHIELD WITH
// NO CHARGES — it never runs out. S30 added the other half: a shield
// that absorbs a fixed amount and then stops.
//
// Both are DATA since ADR 0041 phase 3 tier 3b (#1497): a ScopedEffect
// record with a replacement-reader mod (game/scoped_replacements.go),
// swept by the one duration sweep at cleanup (CR 514.2), carried by
// the snapshot, so a table with a Fog or a half-spent shield on it is
// still a restore point. They used to be closures in a turn-scoped
// registry, which blocked the restore point until cleanup.
//
// The charge is the interesting part, because a charged shield is the
// one continuous effect in the engine that CHANGES as it applies:
// "prevent the next 4 damage" that has already eaten 3 is a different
// effect from the one that was created. The engine never edits the
// record in place — it replaces it with a record holding the charge
// left, and removes a spent one — so an undo, which restores the
// registry a clone took before the damage, rewinds the charge too.
// (The closure it replaced could not: the charge was a captured
// variable an undo did not reach.)

// PreventAllCombatDamageThisTurn is the Fog shape (CR 615.1): every
// combat-damage event for the rest of the turn is cancelled
// outright. No charges, no target — "all combat damage that would
// be dealt this turn" names no object, so the shield fires on every
// damage event there is and simply expires with the turn.
//
// Player narrows it to "… that would be dealt to <Player> this turn"
// (Druid's Deliverance): the controller's blockers still die, and an
// opponent swinging at a third player still gets there.
//
// Non-combat damage is untouched: a Lightning Bolt cast after the
// Fog resolves still kills. That distinction rides on
// ReplacementEvent.IsCombatDamage, set by the combat-damage
// resolver and by nothing else.
type PreventAllCombatDamageThisTurn struct {
	// Player, when set, is the one player whose combat damage is
	// prevented. uuid.Nil is all combat damage.
	Player uuid.UUID

	// Label is the CR 616 ordering-prompt copy, shown when two
	// replacements race for the same event.
	Label string
}

func (p PreventAllCombatDamageThisTurn) Apply(ctx *Context) error {
	label := p.Label
	if label == "" {
		label = "prevent all combat damage this turn"
	}
	ctx.Game.PreventCombatDamageThisTurnForEffect(ctx.Source(), p.Player, label)
	return nil
}

// PreventNextDamage is the CHARGED shield (CR 615.8): "prevent the
// next N damage that would be dealt to <target> this turn".
//
// The arithmetic is per CR 615.8 and is not "cancel if N is big
// enough": a 4-point shield facing 6 damage prevents 4 and lets 2
// through, leaving the shield empty. A 4-point shield facing 3
// damage prevents all 3 and has 1 left for the next event. Getting
// that wrong in the generous direction (cancel the whole event
// whenever the shield covers any of it) is the natural first draft
// and is strictly stronger than printed.
//
// A shield on a PERMANENT is pinned to that object (CR 400.7): it does
// not follow a flicker, and it goes away with the permanent.
//
// Amount must be at least 1. The other printed form — "the next time
// a source would deal damage to you this turn, prevent that damage"
// (the Circle of Protection wording), one use, whole event, whatever
// its size — is not built: no catalogued card prints it, and ADR 0041
// P8 adds a kind only with the first card that needs it. An Amount
// below 1 registers nothing.
type PreventNextDamage struct {
	// Target is the player or permanent being shielded. A card ID
	// and a player ID are the same shape here because
	// ReplacementEvent.DamageTarget is, which is what lets one
	// primitive serve "any target".
	Target uuid.UUID

	// Amount is the damage the shield absorbs, at least 1.
	Amount int

	// CombatOnly narrows the shield to combat damage. Off for every
	// "prevent the next N damage" card, which say nothing about
	// where the damage comes from.
	CombatOnly bool

	Label string
}

func (p PreventNextDamage) Apply(ctx *Context) error {
	if p.Target == uuid.Nil || p.Amount < 1 || ctx.isNewSourceObject(p.Target) { // #1432
		return nil
	}
	label := p.Label
	if label == "" {
		label = "prevent damage"
	}
	ctx.Game.PreventNextDamageThisTurnForEffect(ctx.Source(), p.Target, p.Amount, p.CombatOnly, label)
	return nil
}
