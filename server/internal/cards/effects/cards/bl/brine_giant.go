package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brine Giant — Creature — Giant {6}{U}, 5/6 (Theros Beyond Death, #5514):
//
//	"Affinity for enchantments (This spell costs {1} less to cast for
//	 each enchantment you control.)"
//
// Affinity for enchantments is Sky-Blessed Samurai's AffinityFor cost
// modifier (CR 702.41a): {1} less per enchantment its caster controls,
// never below its colored part.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4ecbaf6a-aa03-45c8-a09e-6c7fc598ab9e",
		Name:         "Brine Giant",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for enchantments", effects.Enchantment()),
		},
	})
}
