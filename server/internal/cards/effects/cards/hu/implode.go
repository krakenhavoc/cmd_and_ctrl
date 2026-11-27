package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Implode — Sorcery {4}{R} (Planeshift, S82 OLD #5615):
//
//	"Destroy target land.
//	 Draw a card."
//
// The destroy, then the draw; a target that became illegal counters the whole
// spell (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7e529372-cec7-40d7-a1ac-1dce1764f8b2",
		Name:         "Implode",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Draws: 1},
		Targets:      effects.TargetPermanent("target land", effects.Land()),
		OnResolve:    effects.OldB637DestroyTargetThenDraw(),
	})
}
