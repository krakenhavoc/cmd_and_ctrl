package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hydrosurge — Instant {U} (Magic 2013, #5568):
//
//	"Target creature gets -5/-0 until end of turn."
//
// Shrink in blue: a layer 7c change (CR 613.4c) that ends at cleanup
// (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4c2059d5-9a2e-4b89-a617-9125eb0cd570",
		Name:         "Hydrosurge",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -5, 0)),
		OnResolve:    effects.BoostTheTargetUntilEOT(-5, 0, "Hydrosurge — -5/-0 until end of turn"),
	})
}
