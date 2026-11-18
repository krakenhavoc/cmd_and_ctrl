package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Harmonious Emergence — Enchantment — Aura {3}{G} (Neon Dynasty, #5493):
//
//	"Enchant land you control
//	 Enchanted land is a 4/5 green Spirit creature with vigilance and
//	 haste. It's still a land.
//	 If enchanted land would be destroyed, instead sacrifice this Aura and
//	 that land gains indestructible until end of turn."
//
// Crackling Emergence's shape (neoB12AnimateAttachedLand and
// neoB12EmergenceReplacement).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ad8fd4e9-f6ae-4f24-b806-4509c8f4ee47",
		Name:         "Harmonious Emergence",
		Completeness: CompletenessFull,
		Targets:      EnchantLand(YouControl()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       neoB12AnimateAttachedLand("G", 4, 5, "vigilance", "haste"),
		Replacements: []game.ReplacementEffect{neoB12EmergenceReplacement()},
	})
}
