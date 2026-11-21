package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aegis of the Heavens — Instant {1}{W} (Core Set 2019, #5527):
//
//	"Target creature gets +1/+7 until end of turn."
//
// A +1/+7 until end of turn on one target, through the shared
// boostTheTargetUntilEOT.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f8e1b38a-2a77-4cf1-8bef-31acb61ede10",
		Name:         "Aegis of the Heavens",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 7)),
		OnResolve:    effects.BoostTheTargetUntilEOT(1, 7, "Aegis of the Heavens — +1/+7"),
	})
}
