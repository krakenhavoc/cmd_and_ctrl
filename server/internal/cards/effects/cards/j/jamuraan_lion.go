package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jamuraan Lion — Creature — Cat {2}{W}, 3/1 (Visions, #5631):
//
//	"{W}, {T}: Target creature can't block this turn."
//
// The {T} keeps it off the turn it arrives (CR 602.5a). The target gains
// the can't-block restriction until cleanup (CR 514.2), which the block
// check reads (CR 509.1b); a target that has left does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b558541f-ae57-41f7-a5bd-e23fa35e68b4",
		Name:         "Jamuraan Lion",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}: Target creature can't block this turn.",
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Purpose: effects.ForTargets(effects.CantBlockUntilEOT(0, 0, 0)),
			Effect:  effects.TargetCreatureCantBlockThisTurn("Jamuraan Lion — target creature can't block this turn"),
		}},
	})
}
