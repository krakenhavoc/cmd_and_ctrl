package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sunbringer's Touch — Sorcery {2}{G}{G} (Dragons of Tarkir, S82 old batch 330):
//
//	"Bolster X, where X is the number of cards in your hand. Each creature
//	 you control with a +1/+1 counter on it gains trample until end of turn.
//	 (To bolster X, choose a creature with the least toughness among creatures
//	 you control and put X +1/+1 counters on it.)"
//
// X is read as the spell resolves, after it has left the hand (CR 608.2h). The
// trample grant is the bolster's tail, so it runs once the counters are on
// (a tie for least toughness is the controller's pick, asked first), and its
// set is fixed then (CR 611.2c): every creature you control that has a +1/+1
// counter on it, whether or not it was the one bolstered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bbfced8f-0f83-4cc9-9143-aa2473828b48",
		Name:         "Sunbringer's Touch",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := effects.CardsInHand(ctx.Game, ctx.Controller())
			if x <= 0 {
				// Bolster 0 chooses a creature and puts nothing on it, then the grant still applies.
				return sunbringersTouchTrample(ctx)
			}
			return effects.Bolster{N: x, ThenTail: ctx.ItemTail(sunbringersTouchTail)}.Apply(ctx)
		},
	})
}

var sunbringersTouchTail = effects.CardTail("old-b330/sunbringers-touch-trample", func(ctx *effects.Context, _ game.TailOutcome) error {
	return sunbringersTouchTrample(ctx)
})

func sunbringersTouchTrample(ctx *effects.Context) error {
	return effects.GrantKeywordUntilEOT{
		Match: effects.And(effects.Creature(), effects.YouControl(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
			return c.Counters[game.CounterPlusOne] > 0
		}),
		Keywords: []string{"trample"},
		Label:    "Sunbringer's Touch — trample until end of turn",
	}.Apply(ctx)
}
