package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aura Blast — Instant {1}{W} (Planeshift, S82 OLD #5615):
//
//	"Destroy target enchantment.
//	 Draw a card."
//
// The destroy, then the draw; a target that became illegal counters the whole
// spell (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4e3c3bdc-667d-42ec-b960-159a39c53cb3",
		Name:         "Aura Blast",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Draws: 1},
		Targets:      effects.TargetPermanent("target enchantment", effects.Enchantment()),
		OnResolve:    effects.OldB637DestroyTargetThenDraw(),
	})
}
