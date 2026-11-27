package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hate Weaver — Creature — Zombie Wizard {1}{B}, 2/1 (Invasion, S82 OLD #5616):
//
//	"{2}: Target blue or red creature gets +1/+0 until end of turn."
//
// A creature is a legal target if it is blue or red (or both) through the
// layers, judged as the ability goes on the stack and again as it resolves
// (CR 608.2b). Its own colour, black, is not a legal one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2a41f61b-a419-4b79-a25f-2506eaf05f55",
		Name:         "Hate Weaver",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: Target blue or red creature gets +1/+0 until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0)),
			Cost:    effects.ManaCost("{2}"),
			Targets: effects.TargetCreature("target blue or red creature", effects.Or(effects.OfColor("U"), effects.OfColor("R"))),
			Effect:  effects.BoostTargetUntilEOT("Hate Weaver — +1/+0 until end of turn", 1, 0),
		}},
	})
}
