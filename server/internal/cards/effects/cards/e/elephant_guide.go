package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elephant Guide — Enchantment — Aura {2}{G} (Eternal Masters, S82 OLD,
// #5542):
//
//	"Enchant creature
//	 Enchanted creature gets +3/+3.
//	 When enchanted creature dies, create a 3/3 green Elephant creature
//	 token."
//
// A layer 7c +3/+3 on the host and Murder Investigation's dies trigger
// (harvested while the Aura is still attached): the Elephant is the Aura
// controller's, whoever controlled the creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c138bd7f-8751-4e96-b54a-d5082e1a491f",
		Name:         "Elephant Guide",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Tokens: 1, CreatureTokens: 1},
		Static:       []game.StaticAbility{effects.PumpEnchantedCreature(3, 3)},
		Triggered: []game.TriggeredAbility{
			effects.WhenEnchantedCreatureDies("Elephant Guide — create a 3/3 green Elephant creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 green Elephant"), N: 1})),
		},
	})
}
