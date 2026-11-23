package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bull Rush — Instant {R} (Worldwake, #5582):
//
//	"Target creature gets +2/+0 until end of turn."
//
// Rouse's pump, without its alternative cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0140cf53-71ef-4d37-9197-48aab35ca4a2",
		Name:         "Bull Rush",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 0)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.BoostTheTargetUntilEOT(2, 0, "Bull Rush — +2/+0"),
	})
}
