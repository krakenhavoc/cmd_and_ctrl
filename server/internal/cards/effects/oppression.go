package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Oppression — Enchantment {1}{B}{B} (EDHREC rank 3193):
//
//	"Whenever a player casts a spell, that player discards a card."
//
// The symmetrical tax: every spell at the table, the controller's
// own included, costs its caster a card. The trigger goes on the
// stack ABOVE the spell that caused it and resolves first, so the
// discard happens with the spell still on the stack — the spell
// itself is never a legal discard, and a caster with an otherwise
// empty hand discards nothing. Oppression's own cast does not
// trigger it (it is not on the battlefield yet). The caster is
// captured off the cast event; the discard is their own choice, and
// until they make it nobody passes priority, so the spell underneath
// cannot resolve out from under the trigger (#651).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f488f679-f5d5-4e61-b0c8-0b0eb622df0d",
		Name:         "Oppression",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor != uuid.Nil
			},
			Key: "Oppression — the spell's caster discards a card",
			Effect: func(g *game.Game, item *game.StackItem) error {
				g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
					Player: item.Trigger.Event.Actor,
					Source: item.SourceCardID,
					N:      1,
				})
				return nil
			},
		}},
	})
}
