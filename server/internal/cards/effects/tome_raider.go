package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tome Raider — Creature — Faerie {2}{U}, 1/1 (Throne of Eldraine, #5516):
//
//	"Flying
//	 When this creature enters, draw a card."
//
// The enters trigger is the shared draw (samiDrawOnETB), so it uses the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ecd9aa9a-06b2-4d8d-9264-fda2036ff142",
		Name:            "Tome Raider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Draws: 1},
		Triggered:       []game.TriggeredAbility{samiDrawOnETB("Tome Raider")},
	})
}
