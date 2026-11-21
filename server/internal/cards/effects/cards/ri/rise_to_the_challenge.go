package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rise to the Challenge — Instant {1}{R} (Born of the Gods,
// #5559):
//
//	"Target creature gets +2/+0 and gains first strike until end of turn."
//
// Interjection's body (socB27TargetGetsAndGains): both halves on the
// target still legal as it resolves, ending at cleanup (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7f3095fb-ffc0-48ed-8792-4157897a9ce7",
		Name:         "Rise to the Challenge",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 0, "first strike")),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.SocB27TargetGetsAndGains(2, 0, "first strike", "Rise to the Challenge"),
	})
}
