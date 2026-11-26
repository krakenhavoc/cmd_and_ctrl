package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Contradict — Instant {3}{U}{U} (Dragons of Tarkir, S82 old batch 329):
//
//	"Counter target spell.
//	 Draw a card."
//
// Spell Snuff's counter with an unconditional draw. The whole spell does
// nothing when its target is gone (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a48f259a-99a0-49c7-a11b-6e796f8a12f4",
		Name:         "Contradict",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Draws: 1},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve:    effects.CounterFirstTargetSpellThenDraw,
	})
}
