package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Evasive Maneuvers — {W} Instant (Star Trek Commander, #4974):
//
//	"Target blocking creature you control gets +2/+2 until end of turn.
//	 Draw a card."
//
// Captain's Defense with "you control" on the target: the same
// resolution (pumpBlockerTwoThenDraw). A creature is blocking only from
// the declaration of blockers to the end of combat (CR 506.4), so the
// spell has no legal target outside that window. If the target is no
// longer legal on resolution the spell does nothing, the draw included
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b2f1364-56e7-461a-b03f-0e14d83b0f5e",
		Name:         "Evasive Maneuvers",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target blocking creature you control", effects.BlockingCreature(), effects.YouControl()),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		OnResolve:    effects.PumpBlockerTwoThenDraw("Evasive Maneuvers — +2/+2"),
	})
}
