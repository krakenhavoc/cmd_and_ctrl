package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shield of the Ages — Artifact {2} (Ice Age, S82 OLD #5636):
//
//	"{2}: Prevent the next 1 damage that would be dealt to you this turn."
//
// A CR 615.7 charged shield on the controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "65164450-b477-4096-8146-ebb80a709787",
		Name:         "Shield of the Ages",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: Prevent the next 1 damage that would be dealt to you this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.ManaCost("{2}"),
			Effect:  effects.PreventNextOneToYouB832,
		}},
	})
}
