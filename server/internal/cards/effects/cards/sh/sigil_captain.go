package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sigil Captain — Creature — Rhino Soldier {1}{G}{W}{W}, 3/3 (Commander 2011,
// S82 older sets batch 454, #5574):
//
//	"Whenever a creature you control enters, if that creature is 1/1, put two
//	 +1/+1 counters on it."
//
// An enters trigger on any creature you control (the Captain included) with an
// intervening "if that creature is 1/1" (CR 603.4): read as it enters and again
// as the ability resolves, from the creature as that object (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7182e6f4-6caf-474a-8759-13c56f391bef",
		Name:         "Sigil Captain",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.S82OldB454OneOneEntered,
				"Sigil Captain — put two +1/+1 counters on that creature", effects.S82OldB454SigilCaptain),
		},
	})
}
