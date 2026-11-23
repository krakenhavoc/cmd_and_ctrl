package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Super Intelligence — {U} Enchantment — Aura (Marvel Super Heroes, #4455):
//
//	"Enchant creature
//	 At the beginning of the upkeep of enchanted creature's controller,
//	 that player draws a card."
//
// The upkeep trigger is Dance of the Dead's (UpkeepOfEnchantedCreaturesController):
// it fires in the upkeep of whoever controls the host now, and the card
// is drawn by that upkeep's player, not by the Aura's controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d82929f2-57ad-45be-a834-984437046026",
		Name:         "Super Intelligence",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, effects.UpkeepOfEnchantedCreaturesController,
				"Super Intelligence — that player draws a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DrawCards{Player: item.Trigger.Event.Actor, N: 1}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
