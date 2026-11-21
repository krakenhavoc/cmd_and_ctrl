package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mortal's Ardor — Instant {W} (Born of the Gods,
// #5559):
//
//	"Target creature gets +1/+1 and gains lifelink until end of turn.
//	 (Damage dealt by the creature also causes its controller to gain
//	 that much life.)"
//
// Interjection's body (socB27TargetGetsAndGains): both halves on the
// target still legal as it resolves, ending at cleanup (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c3401995-4fb4-4b43-b852-5ccad6dc3154",
		Name:         "Mortal's Ardor",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1, "lifelink")),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.SocB27TargetGetsAndGains(1, 1, "lifelink", "Mortal's Ardor"),
	})
}
