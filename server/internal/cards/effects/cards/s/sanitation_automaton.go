package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sanitation Automaton — Artifact Creature — Construct {2}, 2/1 (Murders
// at Karlov Manor, #4469):
//
//	"When this creature enters, surveil 1. (Look at the top card of your
//	 library. You may put it into your graveyard.)"
//
// Nightblade Brigade's entry trigger: the ordinary surveil (CR 701.25a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "be2ab552-df82-49a0-8f0b-ee66de34e1a7",
		Name:         "Sanitation Automaton",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sanitation Automaton — surveil 1", effects.Do(effects.Surveil{N: 1})),
		},
	})
}
