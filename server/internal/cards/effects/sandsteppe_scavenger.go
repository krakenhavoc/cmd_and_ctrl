package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sandsteppe Scavenger — Creature — Dog Scout {4}{G}, 2/2 (Dragons of Tarkir,
// #5551):
//
//	"When this creature enters, bolster 2. (Choose a creature with the least toughness among creatures you control and put two +1/+1 counters on it.)"
//
// Bolster is effects.Bolster (CR 701.39): the creature with the least
// toughness is chosen as the trigger resolves, and does not target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "747ee321-f2dc-4d27-90cd-c4e610086958",
		Name:         "Sandsteppe Scavenger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sandsteppe Scavenger — bolster 2", Do(Bolster{N: 2})),
		},
	})
}
