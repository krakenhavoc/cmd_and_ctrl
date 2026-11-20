package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Devout Monk — Creature — Human Monk Cleric {W}, 1/1 (Starter 1999,
// #5620):
//
//	"When this creature enters, you gain 1 life."
//
// Venerable Monk's enters trigger at 1: on the stack, so it can be
// answered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8a0f4e8e-b541-4327-80fe-4fecc23e0df3",
		Name:         "Devout Monk",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Devout Monk — you gain 1 life", Do(GainLife{Amount: 1})),
		},
	})
}
