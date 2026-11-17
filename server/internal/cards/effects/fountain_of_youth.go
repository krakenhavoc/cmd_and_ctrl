package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fountain of Youth — Artifact {0} (The Dark, #5638):
//
//	"{2}, {T}: You gain 1 life."
//
// Overgrown Arch's life row with a {2} component. The gain uses the
// stack and can be responded to.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b906923c-9997-4da5-a05e-53eb4d2dff32",
		Name:         "Fountain of Youth",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: You gain 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Effect:  gainOneLifeEffect,
		}},
	})
}
