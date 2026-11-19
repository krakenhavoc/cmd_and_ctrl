package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Halving Season — Enchantment {4}{B}:
//
//	"If an opponent would create one or more tokens, they create half that
//	 many of each of those kinds of tokens instead, rounded down.
//	 If an opponent would put one or more counters on a permanent or player,
//	 they put half that many of each of those kinds of counters on that
//	 permanent or player instead, rounded down."
//
// Two replacement effects, the opposite of Doubling Season's. Tokens: each
// group in the creation event (one per kind) is halved and rounded down
// (CR 614.1a, 701.7); a group that rounds to nothing creates none. Counters:
// Vorinclex's halving, an event per kind of counter, by whoever puts them
// (the engine names the placer, CR 122.6); combat damage and a Saga's lore
// counter are not "an opponent putting" counters when the engine does not
// name a placer, and a removal is not a placement.
func init() {
	Register(Spec{
		OracleID:     "4d4ad5b2-b22d-497f-bd9a-282bf5c33947",
		Name:         "Halving Season",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			{
				Watches: []game.EventKind{game.EventTokenCreated},
				AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
					return ev.Kind == game.RepEventCreateTokens && ev.TokenCount() > 0 &&
						ev.TokenController != uuid.Nil && ev.TokenController != src.Controller
				},
				Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
					for i := range ev.TokenGroups {
						if ev.TokenGroups[i].Count > 0 {
							ev.TokenGroups[i].Count /= 2
						}
					}
					return nil
				},
				Controller: vorinclexController,
				Label:      "Halving Season: opponents create half as many tokens",
			},
			{
				Watches: []game.EventKind{game.EventCounterPlaced},
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
					return counterPlacementBy(ev, g, src.Controller, false)
				},
				Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
					ev.CounterDelta /= 2
					return nil
				},
				Controller: vorinclexController,
				Label:      "Halving Season: opponents put half as many counters",
			},
		},
	})
}
