package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Joraga Visionary — Creature — Elf Wizard {3}{G}, 3/2 (Zendikar Rising,
// #5508):
//
//	"When this creature enters, draw a card."
//
// Elvish Visionary's enters trigger, on the stack where it can be
// answered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7854f475-5470-4f2b-99df-af9026895595",
		Name:         "Joraga Visionary",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Joraga Visionary — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
