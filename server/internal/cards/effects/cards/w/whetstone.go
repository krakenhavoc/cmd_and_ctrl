package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Whetstone — Artifact {3} (Urza's Saga, #5624):
//
//	"{3}: Each player mills two cards."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "940e461c-b205-4075-bd1f-a1534c33db6c",
		Name:         "Whetstone",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}: Each player mills two cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{3}"),
			Effect:  effects.S82OldB726EachPlayerMillsTwo,
		}},
	})
}
