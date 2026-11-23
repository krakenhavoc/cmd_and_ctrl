package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crash the Ramparts — Instant {2}{G} (Ixalan, #5532):
//
//	"Target creature gets +3/+3 and gains trample until end of turn."
//
// Fanatical Strength's text: both halves end in the cleanup step
// (CR 514.2), and a target gone by resolution gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8832b76b-58b7-41b2-b2f6-842563b55b36",
		Name:         "Crash the Ramparts",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3, "trample")),
		OnResolve:    effects.SocB27TargetGetsAndGains(3, 3, "trample", "Crash the Ramparts"),
	})
}
