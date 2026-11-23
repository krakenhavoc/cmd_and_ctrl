package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Supply Runners — Creature — Dog {4}{W}, 2/2 (Commander Masters,
// #5948):
//
//	"When this creature enters, put a +1/+1 counter on each other
//	 creature you control."
//
// The set of creatures is fixed as the trigger resolves; Supply Runners
// itself gets none (it says "other").
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4d6da25d-7b62-4a64-bbaa-ae8edab1dd51",
		Name:         "Supply Runners",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Supply Runners — put a +1/+1 counter on each other creature you control",
				effects.PutCounterOnEachOtherCreatureYouControl),
		},
	})
}
