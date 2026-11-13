package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Striped Bears — Creature — Bear {3}{G}, 2/2 (Weatherlight, #5629):
//
//	"When this creature enters, draw a card."
//
// Elvish Visionary's enters trigger: it goes on the stack and the
// controller draws when it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ac856e91-1b88-464b-9420-dc311ce814ea",
		Name:         "Striped Bears",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Striped Bears — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
