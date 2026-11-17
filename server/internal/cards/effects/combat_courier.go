package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Combat Courier — Artifact Creature — Construct {1}, 1/1:
//
//	"{2}, Sacrifice this creature: Draw a card.
//	 Unearth {U}"
func init() {
	Register(Spec{
		OracleID:     "54d32245-c796-4e8e-b61e-c218e3902565",
		Name:         "Combat Courier",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{2}, Sacrifice this creature: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{2}"), SacrificeThis()),
				Effect:  ebDrawACard,
			},
			Unearth("{U}"),
		},
	})
}
