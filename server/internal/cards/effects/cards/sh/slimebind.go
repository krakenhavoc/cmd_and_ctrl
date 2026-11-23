package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slimebind — Enchantment — Aura {1}{U} (Foundations Jumpstart):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature gets -4/-0."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "87aa2cbe-841c-40ff-9b1c-b259b4cc631e",
		Name:            "Slimebind",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraHarm},
		Static:          []game.StaticAbility{effects.PumpAttached(-4, 0)},
	})
}
