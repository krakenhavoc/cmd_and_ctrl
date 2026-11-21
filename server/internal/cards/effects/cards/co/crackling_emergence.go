package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crackling Emergence — Enchantment — Aura {1}{R} (Neon Dynasty, #5493):
//
//	"Enchant land you control
//	 Enchanted land is a 3/3 red Spirit creature with haste. It's still a
//	 land.
//	 If enchanted land would be destroyed, instead sacrifice this Aura and
//	 that land gains indestructible until end of turn."
//
// The animation is one continuous effect across layers 4, 5, 6 and 7b; the
// replacement is neoB12EmergenceReplacement. Shared with Harmonious
// Emergence.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c984787c-f883-42c9-ae5d-9277c84dba01",
		Name:         "Crackling Emergence",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(effects.YouControl()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       effects.NeoB12AnimateAttachedLand("R", 3, 3, "haste"),
		Replacements: []game.ReplacementEffect{effects.NeoB12EmergenceReplacement()},
	})
}
