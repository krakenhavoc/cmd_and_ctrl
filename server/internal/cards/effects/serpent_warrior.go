package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Serpent Warrior — Creature — Snake Warrior {2}{B}, 3/3 (Portal, #5630):
//
//	"When this creature enters, you lose 3 life."
//
// Toxic Abomination's enters trigger, with 3 life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fe249f27-db6c-4826-836e-a04efb1a3eaa",
		Name:         "Serpent Warrior",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Serpent Warrior — you lose 3 life", Do(GainLife{Amount: -3})),
		},
	})
}
