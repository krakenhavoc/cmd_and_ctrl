package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Living Phone — Artifact Creature — Toy {2}{W}, 2/1 (Duskmourn: House
// of Horror, #4466):
//
//	"When this creature dies, look at the top five cards of your library.
//	 You may reveal a creature card with power 2 or less from among them
//	 and put it into your hand. Put the rest on the bottom of your library
//	 in a random order."
//
// Water Tribe Rallier's look-and-take (LookAtTopThenMayTakeToHand) on a
// death trigger: a look, so only you see the five; the card taken is
// revealed; the rest go to the bottom in a random order. A card's power
// in the library is its printed power.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "da554d71-1f13-49ae-9eb2-c778a23df08b",
		Name:         "Living Phone",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Living Phone — look at the top five cards of your library",
				effects.LookAtTopThenMayTakeToHand(5, effects.And(effects.Creature(), effects.PowerLE(2)), 1,
					"Living Phone — you may reveal a creature card with power 2 or less and put it into your hand")),
		},
	})
}
