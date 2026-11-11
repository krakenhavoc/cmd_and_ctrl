package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jacked Rabbit — Creature — Rabbit Warrior {X}{1}{W}, 1/2 (#5368):
//
//	"Ravenous (This creature enters with X +1/+1 counters on it. If X is 5
//	 or more, draw a card when it enters.)
//	 Whenever this creature attacks, create a number of 1/1 white Rabbit
//	 creature tokens equal to this creature's power."
//
// Ravenous is Mawloc's two halves: the CR 614.1c entry counters (XCounters)
// and an enters trigger that draws when the X announced for the spell was 5
// or more (CR 702.156a). The attack trigger counts the Rabbit's power as the
// trigger resolves, or its last-known power if it has left (CR 608.2h); a
// power of zero or less makes none.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                   "5a343a5f-3f9b-43c3-8532-8391de7a5443",
		Name:                       "Jacked Rabbit",
		Completeness:               CompletenessFull,
		XMatters:                   true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(game.CounterPlusOne)},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return Self(ev, source, lki, g) && source.CastX() >= 5
			}, "Jacked Rabbit — ravenous: X was 5 or more, draw a card", Do(DrawCards{N: 1})),
			WheneverThisAttacks("Jacked Rabbit — create a number of 1/1 white Rabbit tokens equal to its power",
				func(g *game.Game, item *game.StackItem) error {
					n := b17LastKnownPowerOffBattlefield(g, item.SourceCardID)
					if onBattlefield(g, item.SourceCardID) {
						if c, ok := g.LookupCardForEffect(item.SourceCardID); ok {
							n = max(c.PowerForComparison(), 0)
						}
					}
					if n <= 0 {
						return nil
					}
					return CreateToken{Template: TokenCard("1/1 white Rabbit"), N: n}.Apply(NewContext(g, item))
				}),
		},
	})
}
