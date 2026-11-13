package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Braidwood Cup — Artifact {3} (Urza's Destiny, #5621):
//
//	"{T}: You gain 1 life."
//
// Overgrown Arch's first row on an artifact. It uses the stack (it is
// not a mana ability), and a noncreature artifact has no summoning
// sickness, so the Cup can be tapped the turn it arrives.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d52b72ff-a82d-430e-94c7-675c83b43e50",
		Name:         "Braidwood Cup",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: You gain 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    TapCost(),
			Effect:  gainOneLifeEffect,
		}},
	})
}
