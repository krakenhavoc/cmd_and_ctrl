package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Snare the Skies — {G} Instant (Avacyn Restored, #5570):
//
//	"Target creature gets +1/+1 and gains reach until end of turn. (It
//	 can block creatures with flying.)"
//
// The shared socB27TargetGetsAndGains body at +1/+1 and reach.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7d011e22-335b-465f-9111-e95a72ebffe9",
		Name:         "Snare the Skies",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1, "reach")),
		OnResolve:    effects.SocB27TargetGetsAndGains(1, 1, "reach", "Snare the Skies"),
	})
}
