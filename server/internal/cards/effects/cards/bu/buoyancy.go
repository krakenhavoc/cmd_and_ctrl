package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Buoyancy — Enchantment — Aura {1}{U} (Mercadian Masques, S82 OLD batch 686, #5619):
//
//	"Flash
//	Enchant creature
//	Enchanted creature has flying."
//
// Flash, and the engine's own flying granted to the creature the Aura is attached to.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f4e4060d-bfff-4991-9ae5-8f848304cd1e",
		Name:            "Buoyancy",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Static:          []game.StaticAbility{effects.GrantToAttached("flying")},
	})
}
