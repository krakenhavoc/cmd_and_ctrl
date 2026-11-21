package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Torch Courier — Creature — Goblin {R}, 1/1 (Jumpstart 2022, #5948):
//
//	"Haste
//	 Sacrifice this creature: Another target creature gains haste until end
//	 of turn."
//
// Akki Drillmaster's haste grant with a sacrifice cost (CR 118.8): the
// Courier is gone before the ability resolves, and "another" is object
// identity.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "34e2d6eb-aad4-4776-872e-cebab3938aa0",
		Name:            "Torch Courier",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: Another target creature gains haste until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "haste")),
			Cost:    effects.SacrificeThis(),
			Targets: effects.Another(effects.TargetCreature("another target creature")),
			Effect:  effects.NeoB07TargetGetsUntilEOT("Torch Courier — haste until end of turn", 0, 0, "haste"),
		}},
	})
}
