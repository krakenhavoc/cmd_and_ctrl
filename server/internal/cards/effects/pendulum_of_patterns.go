package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pendulum of Patterns — Artifact {2} (Aether Revolt,
// #5536):
//
//	"When this artifact enters, you gain 3 life.
//	 {5}, {T}, Sacrifice this artifact: Draw a card."
//
// An enters trigger on the stack and a tap-and-sacrifice draw row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "992075d5-f413-4896-b75e-7bb2d589c50e",
		Name:         "Pendulum of Patterns",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Pendulum of Patterns — you gain 3 life", Do(GainLife{Amount: 3})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{5}, {T}, Sacrifice this artifact: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    Plus(ManaCost("{5}"), TapCost(), SacrificeThis()),
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}
