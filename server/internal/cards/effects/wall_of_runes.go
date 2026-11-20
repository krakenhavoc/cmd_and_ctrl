package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wall of Runes — Creature — Wall {U}, 0/4 (War of the Spark, #5520):
//
//	"Defender (This creature can't attack.)
//	 When this creature enters, scry 1. (Look at the top card of your library. You may put that card on the bottom.)"
//
// Defender is printed data; the enters scry is Glider Kids'.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "35dfe1c7-48ab-436f-9a6e-44835ea01cf9",
		Name:            "Wall of Runes",
		Completeness:    CompletenessFull,
		Purpose:         game.Purpose{Scry: 1},
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Wall of Runes — scry 1", Do(Scry{N: 1})),
		},
	})
}
