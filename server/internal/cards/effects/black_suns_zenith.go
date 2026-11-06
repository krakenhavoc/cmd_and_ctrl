package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Black Sun's Zenith — Sorcery {X}{B}{B} (Avatar: The Last Airbender
// Eternal, #5075):
//
//	"Put X -1/-1 counters on each creature. Shuffle Black Sun's Zenith
//	 into its owner's library."
//
// Every creature on the battlefield as the spell resolves gets X -1/-1
// counters, one placement per creature through the CR 614 counter
// window, put by the caster. The shuffle is Red Sun's Zenith's tail
// (tuck the Zenith, its owner shuffles), run once the last placement has
// settled, so a CR 616 ordering prompt over a counter still comes first.
// With X = 0 nothing is placed and the card is shuffled in all the same.
// A countered Zenith never resolves and goes to the graveyard (the
// ruling).
//
// The same engine-wide note as Red Sun's: if a counter placement pauses
// for a CR 616 choice, the resolution finishes first and the card is
// shuffled into the library from the graveyard once the choice is made.
func init() {
	Register(Spec{
		OracleID:     "038c0165-32b6-4e81-8180-604b49905207",
		Name:         "Black Sun's Zenith",
		XMatters:     true,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, AmountIsX: true}},
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			shuffle := game.KeyedTail{Ref: redSunsZenithShuffle, Player: item.Owner, Source: ctx.Source()}
			var ids []uuid.UUID
			for _, c := range MatchingBattlefield(ctx, Creature()) {
				ids = append(ids, c.InstanceID)
			}
			return avB24PutCountersOnEachThen(ctx, ids, game.CounterMinusOne, ctx.X(), shuffle)
		},
	})
}
