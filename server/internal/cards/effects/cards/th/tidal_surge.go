package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tidal Surge — Sorcery {1}{U} (Portal Second Age, S82 older sets batch 735, #5625):
//
//	"Tap up to three target creatures without flying."
//
// Up to three targets, each without flying as it is chosen and again as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "be738992-77fe-498d-b219-e5da4ce5bf07",
		Name:         "Tidal Surge",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("up to three target creatures without flying", effects.Not(effects.HasKeyword("flying"))).WithCount(0, 3),
		OnResolve:    effects.S82OldB735TapEachTarget,
	})
}
