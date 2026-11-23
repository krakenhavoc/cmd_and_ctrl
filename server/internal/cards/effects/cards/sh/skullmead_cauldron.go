package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skullmead Cauldron — Artifact {4} (Dissension, #5598):
//
//	"{T}: You gain 1 life.
//	 {T}, Discard a card: You gain 3 life."
//
// Two life-gain rows; the discard is a cost, paid before the ability is
// put on the stack (CR 602.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b6864a26-9c13-4c6d-b0e0-d5b5ed9864e8",
		Name:         "Skullmead Cauldron",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: You gain 1 life.",
				Purpose: game.Purpose{LifeGain: 1, Answers: game.AnswerValue},
				Cost:    effects.TapCost(),
				Effect:  effects.GainOneLifeEffect,
			},
			{
				Label:   "{T}, Discard a card: You gain 3 life.",
				Purpose: game.Purpose{LifeGain: 3, Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.TapCost(), effects.DiscardACard()),
				Effect:  effects.Do(effects.GainLife{Amount: 3}),
			},
		},
	})
}
