package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Greel's Caress — Enchantment — Aura {1}{B} (Prophecy, S82 OLD b669, #5617):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature gets -3/-0."
//
// The shared attachment pump (layer 7c), a drawback.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e9e0b78e-07d5-4603-8e3b-27274148d1a1",
		Name:            "Greel's Caress",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraHarm},
		Static:          []game.StaticAbility{effects.PumpAttached(-3, 0)},
	})
}
