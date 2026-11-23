package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Omenspeaker — Creature — Human Wizard {1}{U}, 1/3 (Theros, #5561):
//
//	"When this creature enters, scry 2. (Look at the top two cards of your library, then put any number of them on the bottom and the rest on top in any order.)"
//
// Sage's Row Savant's shape: an enters trigger that scries 2 on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "43d7fcf3-acc2-4e9b-a466-c73b1b6c58af",
		Name:         "Omenspeaker",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Omenspeaker — scry 2", effects.Do(effects.Scry{N: 2})),
		},
	})
}
