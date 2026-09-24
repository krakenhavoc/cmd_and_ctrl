package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// prevention.go — CR 615 damage prevention, the two shapes it comes
// in and the one piece of state that separates them.
//
// S17 shipped Fog as a proof of concept: a turn-scoped replacement
// that cancels every combat-damage event for the rest of the turn.
// That is a SHIELD WITH NO CHARGES — it never runs out, so it needs
// no bookkeeping, and the whole effect fits in a closure that always
// returns Cancel. S30 adds the other half: a shield that absorbs a
// fixed amount and then stops.
//
// The charge is the interesting part, because it is the first
// continuous effect in the engine that MUTATES as it applies. Every
// static ability and every replacement before this one is a pure
// function of board state — recompute from scratch and you get the
// same answer. A prevention shield is not: "prevent the next 4
// damage" that has already eaten 3 is a different object from the
// one that was created, and nothing on the board records the
// difference.
//
// It lives in the closure, as a captured counter. The alternatives
// were worse:
//
//   - A field on Card. The shield does not belong to a permanent —
//     Mending Hands is an instant that goes to the graveyard and
//     leaves the shield behind — and a shield on a PLAYER has no
//     card to hang off at all.
//   - A new Game-level registry. That is a second lifetime to get
//     right (created when? cleared when?) next to one that already
//     has the answer: TurnScopedReplacements is swept at
//     StepCleanup, which is exactly "this turn" (CR 514.2).
//
// The cost of the closure is that undo does not rewind a charge:
// clone.go shares replacement closures rather than copying them, so
// a shield that absorbed damage before an undo is still spent after
// it. That is the long-standing posture for every turn-scoped
// continuous effect — they are already counted in
// ContinuationCensus.TurnScopedReplacements as state a snapshot
// cannot reproduce — and making this one shield the exception would
// mean inventing a serialisable shield type for it alone.

// PreventAllCombatDamageThisTurn is the Fog shape (CR 615.1): every
// combat-damage event for the rest of the turn is cancelled
// outright. No charges, no target — "all combat damage that would
// be dealt this turn" names no object, so the shield fires on every
// damage event there is and simply expires with the turn.
//
// Non-combat damage is untouched: a Lightning Bolt cast after the
// Fog resolves still kills. That distinction rides on
// ReplacementEvent.IsCombatDamage, set by the combat-damage
// resolver and by nothing else.
type PreventAllCombatDamageThisTurn struct {
	// Label is the CR 616 ordering-prompt copy, shown when two
	// replacements race for the same event.
	Label string
}

func (p PreventAllCombatDamageThisTurn) Apply(ctx *Context) error {
	label := p.Label
	if label == "" {
		label = "prevent all combat damage this turn"
	}
	ctx.Game.RegisterTurnScopedReplacement(game.ReplacementEffect{
		Watches: []game.EventKind{game.EventDealDamage},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
			return ev.Kind == game.RepEventDamage && ev.IsCombatDamage
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.Cancel()
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, _ *game.Card) uuid.UUID {
			// CR 616.1 gives the ordering choice to the AFFECTED
			// player — whoever is being dealt the damage — not to
			// the player who registered the shield, so a turn-scoped
			// prevention effect reports no controller. Carried over
			// from S17's Fog verbatim.
			return uuid.Nil
		},
		Label: label,
	})
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
// Amount == 0 is the OTHER printed form — "the next time a source
// would deal damage to you this turn, prevent that damage" (the
// Circle of Protection wording). One use, whole event, regardless
// of size.
type PreventNextDamage struct {
	// Target is the player or permanent being shielded. A card ID
	// and a player ID are the same shape here because
	// ReplacementEvent.DamageTarget is, which is what lets one
	// primitive serve "any target".
	Target uuid.UUID

	// Amount is the damage the shield absorbs. Zero means "the next
	// damage event, whole".
	Amount int

	// CombatOnly narrows the shield to combat damage. Off for every
	// "prevent the next N damage" card, which say nothing about
	// where the damage comes from.
	CombatOnly bool

	Label string
}

func (p PreventNextDamage) Apply(ctx *Context) error {
	if p.Target == uuid.Nil || p.Amount < 0 || ctx.isNewSourceObject(p.Target) { // #1432
		return nil
	}
	label := p.Label
	if label == "" {
		label = "prevent damage"
	}
	target := p.Target
	combatOnly := p.CombatOnly
	// The charge. Captured by both closures below, mutated by
	// Replace, read by AppliesTo — which is how an exhausted shield
	// stops offering itself for the CR 616 ordering prompt instead
	// of applying as a no-op.
	remaining := p.Amount
	wholeEvent := p.Amount == 0
	spent := false

	ctx.Game.RegisterTurnScopedReplacement(game.ReplacementEffect{
		Watches: []game.EventKind{game.EventDealDamage},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
			if ev.Kind != game.RepEventDamage || ev.DamageTarget != target {
				return false
			}
			if combatOnly && !ev.IsCombatDamage {
				return false
			}
			if wholeEvent {
				return !spent
			}
			return remaining > 0
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			if wholeEvent {
				spent = true
				ev.Cancel()
				return nil
			}
			if ev.DamageAmount <= remaining {
				remaining -= ev.DamageAmount
				ev.Cancel()
				return nil
			}
			ev.DamageAmount -= remaining
			remaining = 0
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, _ *game.Card) uuid.UUID {
			return uuid.Nil
		},
		Label: label,
	})
	return nil
}
