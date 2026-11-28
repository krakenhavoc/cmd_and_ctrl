package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Alexi's Cloak — Enchantment — Aura {1}{U} (Prophecy, S82 OLD b669, #5617):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature has shroud."
//
// The shared attachment keyword grant (layer 6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e52eb1a6-fff1-4a47-b434-31a74d76231c",
		Name:            "Alexi's Cloak",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Static:          []game.StaticAbility{effects.GrantToAttached("shroud")},
	})
}
