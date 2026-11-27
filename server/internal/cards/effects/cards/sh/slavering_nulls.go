package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Slavering Nulls — Creature — Goblin Zombie {1}{R} 2/1 (Worldwake, S82 OLD
// batch 489, #5582):
//
//	"Whenever this creature deals combat damage to a player, if you control a
//	 Swamp, you may have that player discard a card."
//
// An intervening "if" (CR 603.4): a Swamp, any permanent with the subtype,
// is checked as the damage is dealt and again as the trigger resolves. The
// discard is optional and the damaged player chooses the card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c6d39445-dfc2-42c3-afd2-016fdf008f0e",
		Name:         "Slavering Nulls",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.S82OldB489WheneverThisDealsCombatDamageToAPlayerIfYouControl("Swamp",
				"Slavering Nulls — you may have that player discard a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DiscardOne(g, item, item.Trigger.Event.Target)
				}), "Slavering Nulls — have that player discard a card?"),
		},
	})
}
