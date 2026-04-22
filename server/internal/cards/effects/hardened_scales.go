package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hardened Scales — "If one or more +1/+1 counters would be put on
// a creature you control, that many plus one +1/+1 counters are
// put on it instead."
//
// Narrower predicate than Doubling Season: +1/+1 counters only,
// creatures only. Bumps CounterDelta by 1 rather than multiplying.
//
// With Doubling Season: order matters. The CR 616 prompt lets the
// affected player (creature's controller) choose. [HS, DS] → (N+1)*2;
// [DS, HS] → (N*2)+1. The sprint exit-criterion test in
// doubling_season_test.go exercises both orders for N=1 (4 vs 3).
//
// AppliesTo gates on:
//   - ev.Kind == RepEventCounter (pre-filter already does this via Watches)
//   - ev.CounterName == "+1/+1"
//   - target is a creature (via Effective().Types lookup — so future
//     type-change effects compose correctly)
//   - target.Controller == source.Controller
func init() {
	Register(Spec{
		OracleID: "a1f3da21-af6d-450e-bf0b-985d158418e6",
		Name:     "Hardened Scales",
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
					ev.CounterDelta += 1
					return nil
				},
				Controller: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) uuid.UUID {
					return src.Controller
				},
				Label: "Hardened Scales: +1 +1/+1 counter",
			},
		},
	})
}
