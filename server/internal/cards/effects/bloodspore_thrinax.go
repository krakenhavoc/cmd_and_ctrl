package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bloodspore Thrinax — Creature — Lizard {2}{G}{G}, 2/2 (Commander 2015,
// #5546):
//
//	"Devour 1 (As this creature enters, you may sacrifice any number of
//	 creatures. It enters with that many +1/+1 counters on it.)
//	 Each other creature you control enters with an additional X +1/+1
//	 counters on it, where X is the number of +1/+1 counters on this
//	 creature."
//
// Devour is the shared as-enters replacement (devour.go, CR 702.82a). The
// second ability is Arwen's CR 614 entry replacement on another creature
// entering under the Thrinax's controller's control, adding the +1/+1
// counters on the Thrinax at that moment; with none it adds nothing.
// Hardened Scales and Doubling Season apply to the counters, since they
// ride the entry event.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1f4743ed-4e91-4344-9409-c8ede5c4804d",
		Name:         "Bloodspore Thrinax",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			Devour("Bloodspore Thrinax", 1),
			{
				Watches:   []game.EventKind{game.EventZoneMove},
				AppliesTo: bnsB13OtherCreatureYouControlEnters,
				Replace: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) error {
					if n := src.Counters[game.CounterPlusOne]; n > 0 {
						ev.AddCounterAtETB(game.CounterPlusOne, n)
					}
					return nil
				},
				Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
					return src.Controller
				},
				Label: "Bloodspore Thrinax: each other creature you control enters with an additional X +1/+1 counters",
			},
		},
	})
}
