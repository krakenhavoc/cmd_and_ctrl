package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dwarven Pony — Creature — Horse {R}, 1/1 (Homelands, S82 OLD b821, #5635):
//
//	"{1}{R}, {T}: Target Dwarf creature gains mountainwalk until end of turn."
//
// A keyword grant to a Dwarf creature, any Dwarf (the Pony is a Horse, so not
// itself).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f47f1c96-0b09-4277-8d61-cffa56d81be7",
		Name:         "Dwarven Pony",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}, {T}: Target Dwarf creature gains mountainwalk until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "mountainwalk")),
			Cost:    effects.Plus(effects.ManaCost("{1}{R}"), effects.TapCost()),
			Targets: effects.TargetCreature("target Dwarf creature", effects.HasSubtype("Dwarf")),
			Effect:  effects.EbTargetCreatureGainsUntilEOT("mountainwalk", "Dwarven Pony — mountainwalk until end of turn"),
		}},
	})
}
