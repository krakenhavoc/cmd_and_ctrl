package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aether Theorist — Creature — Vedalken Rogue {1}{U}, 1/3:
//
//	"When this creature enters, you get {E}{E}{E} (three energy counters).
//	 {T}, Pay {E}: Scry 1. (Look at the top card of your library. You may
//	 put that card on the bottom.)"
//
// ADR 0129 PR 1 (#1995): the energy is a counter on the controller
// (CR 107.14), paid as an activation cost component.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b07ef53e-46db-457e-a5a2-846bb9805b70",
		Name:         "Aether Theorist",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 3},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouGetEnergy("Aether Theorist", 3),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Pay {E}: Scry 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(TapCost(), PayEnergy(1)),
			Effect:  Do(Scry{N: 1}),
		}},
	})
}
