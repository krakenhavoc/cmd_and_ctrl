package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stampede — Instant {1}{G}{G} (Ice Age, S82 OLD, #5636):
//
//	"Attacking creatures get +1/+0 and gain trample until end of turn."
//
// Morale's every-attacker pump with a keyword grant, fixed as it resolves
// (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "49c997d5-7b41-4faf-b898-58db19055671",
		Name:         "Stampede",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		OnResolve:    effects.OldB834Stampede,
	})
}
