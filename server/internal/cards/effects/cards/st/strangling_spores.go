package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Strangling Spores — Instant {3}{B} (Core Set 2019, #5527):
//
//	"Target creature gets -3/-3 until end of turn."
//
// A -3/-3 until end of turn on one target, through the shared
// boostTheTargetUntilEOT.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "444493a6-0f90-4847-9619-4ea8fbf7cf6f",
		Name:         "Strangling Spores",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -3, -3)),
		OnResolve:    effects.BoostTheTargetUntilEOT(-3, -3, "Strangling Spores — -3/-3 until end of turn"),
	})
}
