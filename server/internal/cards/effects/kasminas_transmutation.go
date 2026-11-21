package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kasmina's Transmutation — Enchantment — Aura {1}{U} (Ravnica Remastered):
//
//	"Enchant creature
//	 Enchanted creature loses all abilities and has base power and
//	 toughness 1/1."
//
// Amphibian Downpour's two layers: the ability loss (layer 6) and the
// base-P/T set (layer 7b), in timestamp order with the Aura's own.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a9be16d1-5e53-4b85-8921-c6fe9473f923",
		Name:         "Kasmina's Transmutation",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			LoseAllAbilities(),
			SetAttachedBasePT(1, 1),
		},
	})
}
