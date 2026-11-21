package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skalla Wolf — Creature — Wolf Spirit {3}{G}{G}, 3/3 (Core Set 2019, #5527):
//
//	"When this creature enters, look at the top five cards of your library. You may reveal a green card from among them and put it into your hand. Put the rest on the bottom of your library in a random order."
//
// The enters trigger is LookAtTopThenMayTakeToHand (Horn of the Mark's
// sentence) for a green card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "22da0659-7cd0-4277-b6d2-7d954dcd7729",
		Name:         "Skalla Wolf",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Skalla Wolf — look at the top five cards of your library",
				effects.LookAtTopThenMayTakeToHand(5, effects.OfColor("G"), 1,
					"Skalla Wolf — you may reveal a green card and put it into your hand")),
		},
	})
}
