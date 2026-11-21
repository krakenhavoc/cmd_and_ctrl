package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skycrash — Instant {1}{R} (Aetherdrift, #4464):
//
//	"Destroy target artifact.
//	 Cycling {R} ({R}, Discard this card: Draw a card.)"
//
// The destroy is destroyTheTargetPermanent on any artifact, a Vehicle
// included; cycling is the engine's hand ability (#660).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2343ebff-edf4-4132-ad9e-7800f8014072",
		Name:         "Skycrash",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target artifact", effects.Artifact()),
		OnResolve:    effects.DestroyTheTargetPermanent,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{R}")},
	})
}
