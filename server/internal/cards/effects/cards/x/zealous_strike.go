package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zealous Strike — {1}{W} Instant (Avacyn Restored, #5570):
//
//	"Target creature gets +2/+2 and gains first strike until end of
//	 turn."
//
// Interjection's text, the shared socB27TargetGetsAndGains body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d820ddf4-315d-470a-8d66-663a76a240c6",
		Name:         "Zealous Strike",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2, "first strike")),
		OnResolve:    effects.SocB27TargetGetsAndGains(2, 2, "first strike", "Zealous Strike"),
	})
}
