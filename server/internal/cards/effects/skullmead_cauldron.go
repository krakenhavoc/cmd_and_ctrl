package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "b6864a26-9c13-4c6d-b0e0-d5b5ed9864e8",
		Name:         "Skullmead Cauldron",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{T}: You gain 1 life.",
				Purpose: game.Purpose{LifeGain: 1, Answers: game.AnswerValue},
				Cost:    TapCost(),
				Effect:  gainOneLifeEffect,
			},
			{
				Label:   "{T}, Discard a card: You gain 3 life.",
				Purpose: game.Purpose{LifeGain: 3, Answers: game.AnswerValue},
				Cost:    Plus(TapCost(), DiscardACard()),
				Effect:  Do(GainLife{Amount: 3}),
			},
		},
	})
}
