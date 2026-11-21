package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Indomitable Will — Enchantment — Aura {1}{W} (Foundations Jumpstart):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature gets +1/+2."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cca60afe-b044-4401-8322-170aa015873c",
		Name:            "Indomitable Will",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Static:          []game.StaticAbility{effects.PumpAttached(1, 2)},
	})
}
