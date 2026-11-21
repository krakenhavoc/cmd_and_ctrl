package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Captain's Defense — Instant {W} (Marvel Super Heroes Commander, #4455):
//
//	"Target blocking creature gets +2/+2 until end of turn.
//	 Draw a card."
//
// The target is a blocking creature (CR 506.4). If it is an illegal target
// on resolution the spell does nothing, so no card is drawn (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c3f83115-398f-4738-9bf2-b37bd363e9f6",
		Name:         "Captain's Defense",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target blocking creature", effects.BlockingCreature()),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		OnResolve:    effects.PumpBlockerTwoThenDraw("Captain's Defense — +2/+2"),
	})
}
