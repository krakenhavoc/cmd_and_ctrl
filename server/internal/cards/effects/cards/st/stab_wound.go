package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "aad212b2-57ba-4e70-869b-636842eb7a88",
		Name:         "Stab Wound",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static:       []game.StaticAbility{effects.PumpAttached(-2, -2)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, effects.UpkeepOfEnchantedCreaturesController,
				"Stab Wound — that player loses 2 life",
				func(g *game.Game, item *game.StackItem) error {
					return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Trigger.Event.Actor, -2)
				}),
		},
	})
}
