package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "d82929f2-57ad-45be-a834-984437046026",
		Name:         "Super Intelligence",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginUpkeep, UpkeepOfEnchantedCreaturesController,
				"Super Intelligence — that player draws a card",
				func(g *game.Game, item *game.StackItem) error {
					return DrawCards{Player: item.Trigger.Event.Actor, N: 1}.Apply(NewContext(g, item))
				}),
		},
	})
}
