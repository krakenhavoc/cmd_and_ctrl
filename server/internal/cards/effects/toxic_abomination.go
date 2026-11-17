package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Toxic Abomination — Creature — Phyrexian Zombie {1}{B}, 3/2 (Dominaria United, S82):
//
//	"When this creature enters, you lose 2 life."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6358ee1c-3577-4196-bae4-29a3c160a66a",
		Name:         "Toxic Abomination",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Toxic Abomination — you lose 2 life", Do(GainLife{Amount: -2})),
		},
	})
}
