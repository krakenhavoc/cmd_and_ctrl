package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Doubling Season — "If an effect would create one or more tokens
// under your control, it creates twice that many of those tokens
// instead. If an effect would put one or more counters on a
// permanent you control, it puts twice that many of those counters
// on that permanent instead."
//
// First S17 catalog card with a replacement effect. Ships only the
// counter-doubling half; token-doubling lands with S21 (tokens,
// sacrifice, aristocrats) when the token primitive exists.
//
// How the pipeline sees it:
//   - Watches EventCounterPlaced (pre-event, via
//     game.RepEventCounter).
//   - AppliesTo: the counter target is a permanent the Season's
//     controller controls. Any counter name (+1/+1, loyalty, charge,
//     etc.) — "counters" in CR is uncategorised.
//   - Replace: ev.CounterDelta *= 2.
//
// Multi-replacement ordering (CR 616): with Hardened Scales also in
// play, the affected player (counter target's controller) picks the
// order — [HS, DS] → (1+1)*2 = 4; [DS, HS] → (1*2)+1 = 3. The CR
// 616 order prompt queues; the exit-criterion test asserts both
// results in effects/doubling_season_test.go.
func init() {
	Register(Spec{
		OracleID: "01546b7d-a233-4176-8843-d732074dc5b6",
		Name:     "Doubling Season",
		Replacements: []game.ReplacementEffect{
			{
				Watches: []game.EventKind{game.EventCounterPlaced},
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
					if ev.Kind != game.RepEventCounter {
						return false
					}
					target, ok := g.LookupCardForEffect(ev.CounterTarget)
					if !ok {
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
				Label: "Doubling Season: double counters",
			},
		},
	})
}
