package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Messenger Falcons — Creature — Bird {2}{G/U}{W}, 2/2 (Alara Reborn,
// #5586):
//
//	"Flying
//	 When this creature enters, draw a card."
//
// The enters trigger uses the stack (Arctic Wolves' draw).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "01fda82a-227d-4e3d-8481-a13c3812392a",
		Name:            "Messenger Falcons",
		Completeness:    CompletenessFull,
		Purpose:         game.Purpose{Draws: 1},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Messenger Falcons — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
