package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Teroh's Faithful — Creature — Human Cleric {3}{W}, 1/4 (Torment,
// #5612):
//
//	"When this creature enters, you gain 4 life."
//
// Migrating Ketradon's enters trigger without the reach and the cycling:
// a real trigger on the stack, so it can be answered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "875c96ce-f6e4-438b-929d-de53afdd99a1",
		Name:         "Teroh's Faithful",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{LifeGain: 4},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Teroh's Faithful — you gain 4 life", Do(GainLife{Amount: 4})),
		},
	})
}
