package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Toxic Abomination — Creature — Phyrexian Zombie {1}{B}, 3/2 (Dominaria United, S82):
//
//	"When this creature enters, you lose 2 life."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6358ee1c-3577-4196-bae4-29a3c160a66a",
		Name:         "Toxic Abomination",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Toxic Abomination — you lose 2 life", effects.Do(effects.GainLife{Amount: -2})),
		},
	})
}
