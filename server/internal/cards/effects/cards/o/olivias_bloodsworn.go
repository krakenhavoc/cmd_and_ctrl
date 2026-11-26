package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Olivia's Bloodsworn — Creature — Vampire Soldier {1}{B}, 2/1 (Shadows over
// Innistrad, S82 OLD):
//
//	"Flying
//	 This creature can't block.
//	 {R}: Target Vampire gains haste until end of turn."
//
// Aesthir Glider's can't-block restriction and Air Marshal's grant over a
// Vampire target, the Bloodsworn itself included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e2cfbaf5-903c-49fa-b7bf-2c55087f1ffa",
		Name:            "Olivia's Bloodsworn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: Target Vampire gains haste until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "haste")),
			Cost:    effects.ManaCost("{R}"),
			Targets: effects.TargetCreature("target Vampire", effects.OfCreatureType("Vampire")),
			Effect:  effects.EbTargetCreatureGainsUntilEOT("haste", "Olivia's Bloodsworn — haste until end of turn"),
		}},
	})
}
