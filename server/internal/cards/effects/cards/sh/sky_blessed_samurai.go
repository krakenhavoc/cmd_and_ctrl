package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sky-Blessed Samurai — Enchantment Creature — Human Samurai {6}{W}, 4/4
// (Kamigawa: Neon Dynasty, #5493):
//
//	"Affinity for enchantments (This spell costs {1} less to cast for each
//	 enchantment you control.)
//	 Flying"
//
// Affinity is Assert Authority's self cost modifier over enchantments.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dad767d4-e7d4-4fe0-8a18-bb434c5268eb",
		Name:            "Sky-Blessed Samurai",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for enchantments", effects.Enchantment()),
		},
	})
}
