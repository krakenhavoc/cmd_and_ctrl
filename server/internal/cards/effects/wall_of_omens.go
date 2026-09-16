package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wall of Omens — Creature — Wall {1}{W}, 0/4 (EDHREC rank 1098):
//
//	"Defender
//	 When this creature enters, draw a card."
//
// The white cantrip wall. Defender is the printed keyword the combat
// engine reads; the draw is a real ETB trigger with a response
// window.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5f601f48-d24b-4883-9fde-b3f620e7c9ea",
		Name:            "Wall of Omens",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Wall of Omens — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
