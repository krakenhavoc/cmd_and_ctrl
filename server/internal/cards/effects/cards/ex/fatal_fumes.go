package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fatal Fumes — Instant {3}{B}:
//
//	"Target creature gets -4/-2 until end of turn."
//
// The shared boostTheTargetUntilEOT body at -4/-2, ending at cleanup
// (CR 514.2), declared as a harm so it is aimed at an opponent's creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "359ceca4-7063-4089-8088-ba54bc6fb7cf",
		Name:         "Fatal Fumes",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -4, -2)),
		OnResolve:    effects.BoostTheTargetUntilEOT(-4, -2, "Fatal Fumes — -4/-2 until end of turn"),
	})
}
