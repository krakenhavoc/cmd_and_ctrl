package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eyetwitch — Creature — Eye Bat {B}, 1/1 (EDHREC rank 15744):
//
//	"Flying
//	 When this creature dies, learn. (You may reveal a Lesson card you
//	 own from outside the game and put it into your hand, or discard a
//	 card to draw a card.)"
//
// A dies trigger whose instruction is learn (CR 701.48a, learn.go): the
// trigger's controller is the one who learns.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6a7a5e93-fea8-4aa0-a2b1-73d466220cfb",
		Name:            "Eyetwitch",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Eyetwitch — learn", Do(Learn{})),
		},
	})
}
