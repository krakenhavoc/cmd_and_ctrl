package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thunder Strike — Instant {1}{R} (Magic 2011, #5579):
//
//	"Target creature gets +2/+0 and gains first strike until end of
//	 turn."
//
// Fit of Rage's pump and grant at +2/+0: a layer 7c change and a layer 6
// grant that both end at cleanup.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a5f25e80-7a6a-4629-8201-901e0fcd4a02",
		Name:         "Thunder Strike",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 0, "first strike")),
		OnResolve:    effects.SocB27TargetGetsAndGains(2, 0, "first strike", "Thunder Strike"),
	})
}
