package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "dad767d4-e7d4-4fe0-8a18-bb434c5268eb",
		Name:            "Sky-Blessed Samurai",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for enchantments", Enchantment()),
		},
	})
}
