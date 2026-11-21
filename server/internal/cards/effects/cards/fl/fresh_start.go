package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fresh Start — Enchantment — Aura {1}{U} (Tarkir: Dragonstorm, #4463):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature gets -5/-0 and loses all abilities."
//
// Quantum Reduction's statics (PumpAttached and LoseAllAbilities) with
// printed flash: they last exactly as long as the Aura is attached.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "77f1734d-ea0b-4344-8a9e-9d1a34eb6164",
		Name:            "Fresh Start",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.PumpAttached(-5, 0),
			effects.LoseAllAbilities(),
		},
	})
}
