package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mortal's Resolve — Instant {1}{G} (Born of the Gods,
// #5559):
//
//	"Target creature gets +1/+1 and gains indestructible until end of
//	 turn. (Damage and effects that say "destroy" don't destroy it.)"
//
// Interjection's body (socB27TargetGetsAndGains): both halves on the
// target still legal as it resolves, ending at cleanup (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ef29f979-1a6d-47bb-83f5-4679389471da",
		Name:         "Mortal's Resolve",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1, "indestructible")),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.SocB27TargetGetsAndGains(1, 1, "indestructible", "Mortal's Resolve"),
	})
}
