package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Megrim — Enchantment {2}{B} (EDHREC rank 2584):
//
//	"Whenever an opponent discards a card, this enchantment deals 2
//	 damage to that player."
//
// The original discard payoff. Liliana's Caress's condition — the
// discarding player is the event's Actor (b18OpponentDiscarded) —
// with DAMAGE from Megrim to that player rather than life loss, so a
// prevention shield or a damage doubler sees it. Once per card
// discarded, so a wheel that makes an opponent discard seven is
// fourteen damage, as printed. The player is captured in Build.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "633ad9e2-9f55-4a1c-9248-661ad4b0e1dc",
		Name:         "Megrim",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDiscardCard},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.B18OpponentDiscarded(ev, source)
			},
			Key: "Megrim — deal 2 damage to that player",
			Effect: func(g *game.Game, item *game.StackItem) error {
				victim := item.Trigger.Event.Actor
				return effects.DealDamage{Source: item.SourceCardID, Target: victim, Amount: 2}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
