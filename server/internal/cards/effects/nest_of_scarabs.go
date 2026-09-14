package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nest of Scarabs — Enchantment {2}{B} (EDHREC rank 3495):
//
//	"Whenever you put one or more -1/-1 counters on a creature, create
//	 that many 1/1 black Insect creature tokens."
//
// The -1/-1 counters deck's payoff. The engine emits one
// EventCounterPlaced per placement, carrying the post-change total,
// so "one or more" is one event and "that many" is its delta — the
// previous total read back off the log (b33CountersPlacedDelta), the
// way b11CountersWerePlaced tells a placement from a removal. "You
// put" is b11ResolvingController: counters land during a resolution,
// and the player resolving is the player putting them, so an
// opponent's Black Sun's Zenith on the controller's creature makes
// the opponent's Nest's Insects, not this one's. Any creature — the
// controller's own, an opponent's — counts, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1f21cf59-6390-44b4-ab2e-ef290dfd8c85",
		Name:         "Nest of Scarabs",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCounterPlaced},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b33YouPutMinusCountersOnACreature(ev, source, g)
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				n := b33CountersPlacedDelta(ev, game.CounterMinusOne, g)
				return game.NewTriggeredItem(source, "Nest of Scarabs — create a 1/1 black Insect for each -1/-1 counter placed",
					b33CreateInsectsPerMinusCounterPlaced(n))
			},
		}},
	})
}
