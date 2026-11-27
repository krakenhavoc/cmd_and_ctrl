package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stone Giant — Creature — Giant {2}{R}{R}, 3/4 (Magic 2010, #5585):
//
//	"{T}: Target creature you control with toughness less than this
//	 creature's power gains flying until end of turn. Destroy that creature
//	 at the beginning of the next end step."
//
// The target is relative to the Giant (Unliving Psychopath's reading, with
// the candidate's toughness against the Giant's power: the comparison is
// made as the ability is announced and again as it resolves, against
// last-known power, CR 608.2h). The creature gains flying until end of turn
// and a delayed trigger destroys it at the beginning of the next end step
// if it is still that object (CR 603.7c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0b8e3f9b-a4da-49a3-8545-ce7a265e5856",
		Name:         "Stone Giant",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target creature you control with toughness less than this creature's power gains flying until end of turn. Destroy that creature at the beginning of the next end step.",
			Cost:    effects.TapCost(),
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying")),
			Targets: effects.RelativeToSource(
				effects.TargetCreature("target creature you control with toughness less than this creature's power", effects.YouControl()),
				effects.OldB499ToughnessLessThanPower()),
			Effect: effects.OldB499FlyThenDestroyAtEndStep("Stone Giant"),
		}},
	})
}
