package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Novice Inspector — Creature — Human Detective {W}, 1/2 (Murders at
// Karlov Manor, #4469):
//
//	"When this creature enters, investigate. (Create a Clue token. It's
//	 an artifact with "{2}, Sacrifice this token: Draw a card.")"
//
// Thraben Inspector under a new name: one Clue for its controller
// (CR 701.16a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c206bdcb-d2c4-4bda-a979-6c0627ef8f73",
		Name:         "Novice Inspector",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Novice Inspector — investigate", Do(Investigate{N: 1})),
		},
	})
}
