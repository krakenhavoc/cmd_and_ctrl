package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spirit Weaver — Creature — Human Wizard {1}{W}, 2/1 (Invasion, S82 OLD
// #5616):
//
//	"{2}: Target green or blue creature gets +0/+1 until end of turn."
//
// Hate Weaver's shape: a creature is a legal target if it is green or blue (or
// both) through the layers, judged as the ability goes on the stack and again
// as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4e6e28f1-111b-4bac-b814-487f6c206485",
		Name:         "Spirit Weaver",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: Target green or blue creature gets +0/+1 until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 1)),
			Cost:    effects.ManaCost("{2}"),
			Targets: effects.TargetCreature("target green or blue creature", effects.Or(effects.OfColor("G"), effects.OfColor("U"))),
			Effect:  effects.BoostTargetUntilEOT("Spirit Weaver — +0/+1 until end of turn", 0, 1),
		}},
	})
}
