package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stab Wound — Enchantment — Aura {2}{B} (Enchanting Tales):
//
//	"Enchant creature
//	 Enchanted creature gets -2/-2.
//	 At the beginning of the upkeep of enchanted creature's controller, that
//	 player loses 2 life."
//
// Dead Weight's -2/-2 and Super Intelligence's upkeep trigger: it fires in
// the upkeep of whoever controls the host now, and the life is lost by that
// upkeep's player, not by the Aura's controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aad212b2-57ba-4e70-869b-636842eb7a88",
		Name:         "Stab Wound",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static:       []game.StaticAbility{PumpAttached(-2, -2)},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginUpkeep, UpkeepOfEnchantedCreaturesController,
				"Stab Wound — that player loses 2 life",
				func(g *game.Game, item *game.StackItem) error {
					return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Trigger.Event.Actor, -2)
				}),
		},
	})
}
