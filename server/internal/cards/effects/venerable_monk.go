package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Venerable Monk — Creature — Human Monk Cleric {2}{W}, 2/2 (Portal, #5630):
//
//	"When this creature enters, you gain 2 life."
//
// An enters trigger on the stack, like Mandroid Squadron's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5c5cfa6d-857f-44a9-80f7-46f016ef71e4",
		Name:         "Venerable Monk",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Venerable Monk — you gain 2 life", Do(GainLife{Amount: 2})),
		},
	})
}
