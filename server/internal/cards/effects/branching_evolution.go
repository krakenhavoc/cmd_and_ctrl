package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Branching Evolution — "If one or more +1/+1 counters would be
// put on a creature you control, twice that many +1/+1 counters
// are put on it instead."
//
// Same predicate shape as Hardened Scales (+1/+1 only, creatures
// only, own permanents) but multiplies rather than adds.
// Architecturally a narrower Doubling Season.
//
// All three of Doubling Season / Hardened Scales / Branching
// Evolution on the battlefield together → three-way CR 616 prompt.
// Order: [HS, BE, DS] for 1 counter → (1+1)*2*2 = 8;
// [DS, BE, HS] → (1*2)*2+1 = 5. The test scenarios don't cover
// every permutation, but the engine's apply-loop + prompt shape
// handle N replacements uniformly.
func init() {
	Register(Spec{
		OracleID: "28fe909b-06e0-424c-9f75-c824a25f5865",
		Name:     "Branching Evolution",
		Replacements: []game.ReplacementEffect{
			{
				Watches: []game.EventKind{game.EventCounterPlaced},
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
					if ev.Kind != game.RepEventCounter {
						return false
					}
					if ev.CounterName != "+1/+1" {
						return false
					}
					target, ok := g.LookupCardForEffect(ev.CounterTarget)
					if !ok {
						return false
					}
					if !target.IsCreature() {
						return false
					}
					return target.Controller == src.Controller
				},
				Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
					ev.CounterDelta *= 2
					return nil
				},
				Controller: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) uuid.UUID {
					return src.Controller
				},
				Label: "Branching Evolution: double +1/+1 counters",
			},
		},
	})
}
