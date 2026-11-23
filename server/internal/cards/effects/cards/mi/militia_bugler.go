package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Militia Bugler — Creature — Human Soldier {2}{W}, 2/3
// (Core Set 2019, #5527):
//
//	"Vigilance (Attacking doesn't cause this creature to tap.)
//	When this creature enters, look at the top four cards of your library. You may reveal a creature card with power 2 or less from among them and put it into your hand. Put the rest on the bottom of your library in a random order."
//
// Vigilance, and Squad Rallier's look-at-four as an enters trigger
// (LookAtTopThenMayTakeToHand).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "edb6ae9e-a41c-4d82-97bf-4be00e98364c",
		Name:            "Militia Bugler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Militia Bugler — look at the top four cards of your library",
				effects.LookAtTopThenMayTakeToHand(4, effects.And(effects.Creature(), effects.PowerLE(2)), 1,
					"Militia Bugler — you may reveal a creature card with power 2 or less and put it into your hand")),
		},
	})
}
