package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arashin Cleric — Creature — Human Cleric {1}{W}, 1/3 (Fate Reforged,
// #5552):
//
//	"When this creature enters, you gain 3 life."
//
// Temple Acolyte's entry trigger: it uses the stack, so it can be answered,
// and the life is gained as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "69bcdb42-c170-4848-b95b-f4621867d084",
		Name:         "Arashin Cleric",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{LifeGain: 3},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Arashin Cleric — gain 3 life", Do(GainLife{Amount: 3})),
		},
	})
}
