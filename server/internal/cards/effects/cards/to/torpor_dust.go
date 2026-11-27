package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Torpor Dust — Enchantment — Aura {2}{U/B} (Shadowmoor, S82 older sets):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature gets -3/-0."
//
// Way of the Thief's attached pump with a negative power, and the flash keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "da1ca3ec-8bef-4b7e-896e-4d50e8cffbe5",
		Name:            "Torpor Dust",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraHarm},
		Static:          []game.StaticAbility{effects.PumpAttached(-3, 0)},
	})
}
