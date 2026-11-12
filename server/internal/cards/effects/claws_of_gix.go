package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Claws of Gix — Artifact {0} (Urza's Saga):
//
//	"{1}, Sacrifice a permanent: You gain 1 life."
//
// A sacrifice outlet for any permanent you control, the Claws included.
// It uses the stack (it is not a mana ability), and the sacrifice is a
// cost, so it happens even if the ability is countered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c4d384d7-f294-4b2d-9971-a4689c150255",
		Name:         "Claws of Gix",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice a permanent: You gain 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{1}"), SacrificeAPermanent()),
			Effect:  Do(GainLife{Amount: 1}),
		}},
	})
}
