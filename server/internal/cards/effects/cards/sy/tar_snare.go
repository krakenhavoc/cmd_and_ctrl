package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tar Snare — Instant {2}{B} (Oath of the Gatewatch, #5545):
//
//	"Target creature gets -3/-2 until end of turn."
//
// The shared boostTheTargetUntilEOT with negative amounts, as Disfigure.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3621bd8c-c883-462f-ad56-c01deecb2720",
		Name:         "Tar Snare",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -3, -2)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.BoostTheTargetUntilEOT(-3, -2, "Tar Snare — -3/-2"),
	})
}
