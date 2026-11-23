package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rush of Adrenaline — Instant {R} (Shadows over Innistrad, #5543):
//
//	"Target creature gets +2/+1 and gains trample until end of turn."
//
// Interjection's pump-and-keyword body (socB27TargetGetsAndGains) at +2/+1
// and trample.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ec9975f4-dfff-4b72-9da8-8f2025f0524a",
		Name:         "Rush of Adrenaline",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 1, "trample")),
		OnResolve:    effects.SocB27TargetGetsAndGains(2, 1, "trample", "Rush of Adrenaline"),
	})
}
