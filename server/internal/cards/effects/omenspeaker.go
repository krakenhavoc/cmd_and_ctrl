package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Omenspeaker — Creature — Human Wizard {1}{U}, 1/3 (Theros, #5561):
//
//	"When this creature enters, scry 2. (Look at the top two cards of your library, then put any number of them on the bottom and the rest on top in any order.)"
//
// Sage's Row Savant's shape: an enters trigger that scries 2 on the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "43d7fcf3-acc2-4e9b-a466-c73b1b6c58af",
		Name:         "Omenspeaker",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Scry: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Omenspeaker — scry 2", Do(Scry{N: 2})),
		},
	})
}
