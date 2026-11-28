package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blinking Spirit — Creature — Spirit {3}{W}, 2/2 (Ice Age, S82 OLD, #5636):
//
//	"{0}: Return this creature to its owner's hand."
//
// Ghost Town's free self-bounce on a creature: it uses the stack, so it can be
// answered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a26f6958-9bc8-46f8-8ca7-7f55098e1291",
		Name:         "Blinking Spirit",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{0}: Return this creature to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{0}"),
			Effect:  effects.ReturnThisPermanentToOwnersHand,
		}},
	})
}
