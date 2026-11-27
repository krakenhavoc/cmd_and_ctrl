package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mageta's Boon — Enchantment — Aura {1}{W} (Prophecy, S82 OLD b669, #5617):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature gets +1/+2."
//
// The shared attachment pump (layer 7c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e3e5b12c-2103-4b40-83d8-6d5449179b6f",
		Name:            "Mageta's Boon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Static:          []game.StaticAbility{effects.PumpAttached(1, 2)},
	})
}
