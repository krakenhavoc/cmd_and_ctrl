package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wall of Blossoms — Creature — Plant Wall {1}{G}, 0/4 (EDHREC rank
// 2218):
//
//	"Defender
//	 When this creature enters, draw a card."
//
// The cantrip wall. Defender rides PrintedKeywords; the ETB is a
// mandatory draw on the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ef4d5fb3-70a3-433d-a9d3-18b2beb8d79f",
		Name:            "Wall of Blossoms",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Wall of Blossoms — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
