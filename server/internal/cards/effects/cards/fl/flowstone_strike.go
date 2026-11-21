package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flowstone Strike — Instant {1}{R}
// (Nemesis, #5618):
//
//	"Target creature gets +1/-1 and gains haste until end of turn."
//
// Interjection's body (socB27TargetGetsAndGains) at +1/-1 and haste.
// Both end in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d67048eb-eaf1-4b5b-a8e3-a004bab438f3",
		Name:         "Flowstone Strike",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "haste")),
		OnResolve:    effects.SocB27TargetGetsAndGains(1, -1, "haste", "Flowstone Strike"),
	})
}
