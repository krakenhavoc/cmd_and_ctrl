package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Parasitic Bond — Enchantment — Aura {3}{B} (Urza's Saga, #5624):
//
//	"Enchant creature
//	 At the beginning of the upkeep of enchanted creature's controller, this
//	 Aura deals 2 damage to that player."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9d973df5-b2e8-4af8-8379-0f8282e1e62a",
		Name:         "Parasitic Bond",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, effects.UpkeepOfEnchantedCreaturesController,
				"Parasitic Bond — 2 damage to that player",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DealDamage{Source: item.SourceCardID, Target: item.Trigger.Event.Actor, Amount: 2}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
