package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Feral Invocation — Enchantment — Aura {2}{G} (Theros, #5561):
//
//	"Flash (You may cast this spell any time you could cast an instant.)
//	 Enchant creature
//	 Enchanted creature gets +2/+2."
//
// A flash Aura with a static pump.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5b3c069b-5bed-4b70-b04c-b7143a4a6e96",
		Name:            "Feral Invocation",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Static:          []game.StaticAbility{effects.PumpAttached(2, 2)},
	})
}
