package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silk Net — Instant {G} (Urza's Legacy, S82 OLD b715, #5623):
//
//	"Target creature gets +1/+1 and gains reach until end of turn. (It can block creatures with flying.)"
//
// Snare the Skies' text: the shared SocB27TargetGetsAndGains body at +1/+1
// and reach.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "58511cb8-348d-409e-8fd6-772be20e57cd",
		Name:         "Silk Net",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1, "reach")),
		OnResolve:    effects.SocB27TargetGetsAndGains(1, 1, "reach", "Silk Net"),
	})
}
