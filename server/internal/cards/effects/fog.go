package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fog — "Prevent all combat damage that would be dealt this turn."
//
// S17 sub-PR 5 proof-of-concept for damage prevention via the CR
// 614 replacement pipeline. Turn-scoped (not battlefield-presence-
// gated) because Fog is an instant — it resolves, registers a
// transient replacement effect, and goes to the graveyard. The
// pipeline keeps firing the registered replacement on every
// MarkCombatDamage call until StepCleanup clears the turn-scoped
// registry.
//
// The replacement fires on any RepEventDamage with
// IsCombatDamage=true and cancels the event. Non-combat damage
// (Lightning Bolt after Fog resolves) is untouched — the
// IsCombatDamage flag comes from the combat-damage resolver's
// markCombatDamageOnCardLocked / markCombatDamageToPlayerLocked
// calls; spell damage via DealDamageToCreatureForEffect doesn't
// set it.
//
// Full CR 615 damage-prevention shields with charges (Shield of
// the Oversoul, Story Circle — stateful "N uses remaining") ships
// in S30. S17 Fog is the simplest atomic cancel, scaffolded so S30
// can layer the charge mechanic on top.
//
// Registration happens in OnResolve rather than via Spec.Replacements
// because Fog is an instant that goes to the graveyard after
// resolving. Its effect has to survive past the source card
// leaving the stack.
func init() {
	Register(Spec{
		OracleID: "27e9db49-7af7-4bef-ad4c-bf5dfb92030d",
		Name:     "Fog",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			// Capture the controller and current turn number at
			// registration time. A future "skip your next turn" or
			// multi-turn Fog would key on a specific turn number;
			// vanilla Fog's turn-scoped cleanup handles expiry.
			_ = ctx.Controller()
			ctx.Game.RegisterTurnScopedReplacement(game.ReplacementEffect{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
					return ev.Kind == game.RepEventDamage && ev.IsCombatDamage
				},
				Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
					ev.Cancel()
					return nil
				},
				Controller: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) uuid.UUID {
					// Turn-scoped replacements have no source card; the
					// controller is whoever cast the spell. S17 returns
					// uuid.Nil because the CR 616 prompt needs a chooser
					// only when ≥2 replacements apply — and for combat
					// damage prevention chaining, CR 616 says the
					// affected player (damage target's controller)
					// picks order, not the registering spell's
					// controller.
					return uuid.Nil
				},
				Label: "Fog: prevent combat damage",
			})
			return nil
		},
	})
}
