package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drag Under — Sorcery {2}{U} (Foundations Jumpstart):
//
//	"Return target creature to its owner's hand.
//	 Draw a card."
//
// Repulse's body at sorcery speed (bounceTargetThenDrawACard).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "021d8596-782a-4cfd-8b70-ac5bb8a5aff1",
		Name:         "Drag Under",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.BounceTargetThenDrawACard,
	})
}
