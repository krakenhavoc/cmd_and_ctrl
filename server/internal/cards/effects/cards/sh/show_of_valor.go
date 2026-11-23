package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Show of Valor — Instant {1}{W} (Magic 2013, #5568):
//
//	"Target creature gets +2/+4 until end of turn."
//
// A layer 7c pump (CR 613.4c) on the target, ending at cleanup.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9645c1cb-2305-4bc0-89d0-a50815e91573",
		Name:         "Show of Valor",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 4)),
		OnResolve:    effects.BoostTheTargetUntilEOT(2, 4, "Show of Valor — +2/+4 until end of turn"),
	})
}
