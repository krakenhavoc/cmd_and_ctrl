package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Disorient — Instant {3}{U} (Magic 2010, #5585):
//
//	"Target creature gets -7/-0 until end of turn."
//
// Shrink's shape with -7/-0: a layer 7c change (CR 613.4c) that ends at
// cleanup (CR 514.2). If the target has left, the spell does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "33be346b-a232-4680-9b1b-a5eee57d2e75",
		Name:         "Disorient",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -7, 0)),
		OnResolve:    effects.BoostTheTargetUntilEOT(-7, 0, "Disorient — -7/-0 until end of turn"),
	})
}
