package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jawbone Skulkin — Artifact Creature — Scarecrow {1}, 1/1 (Eventide, #5590):
//
//	"{2}: Target red creature gains haste until end of turn."
//
// The colour is read off the target as the ability is announced and again
// as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1cbcc073-1645-4d7f-98f4-7309d7b35692",
		Name:         "Jawbone Skulkin",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: Target red creature gains haste until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "haste")),
			Cost:    effects.ManaCost("{2}"),
			Targets: effects.TargetCreature("target red creature", effects.OfColor("R")),
			Effect:  effects.EbTargetCreatureGainsUntilEOT("haste", "Jawbone Skulkin — haste until end of turn"),
		}},
	})
}
