package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s40_helpers.go — card-side helpers of batch 40 (suspend, foretell,
// plot, miracle and casting from outside the hand).

// cmdr2S40ExileWithTimeCounters is "Exile <this spell> with N time counters
// on it": the resolving spell exiles itself and the card that really
// reached exile gets the counters, so it is suspended (CR 702.62b).
func cmdr2S40ExileWithTimeCounters(ctx *Context, n int) error {
	return ExileTarget{Target: ctx.Item.SourceCardID, ThenTail: ctx.TailN(cmdr2S40TimeCounters, n)}.Apply(ctx)
}

// cmdr2S40TimeCounters puts Tail.N time counters on each card that reached
// exile. An on-disk key.
var cmdr2S40TimeCounters = CardTail("cmdr2-s40/time-counters", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if err := ctx.Game.AddCounterByForEffect(ctx.Controller(), id, game.CounterTime, o.Tail.N); err != nil {
			return err
		}
	}
	return nil
})

// cmdr2S40CastFromOutsideHandCostsLess is "Spells you cast from anywhere
// other than your hand cost {N} less to cast" (Savvy Trader, Sage of the
// Beyond): a commander from the command zone, a foretold, plotted or
// suspended card from exile, a flashback card from the graveyard.
func cmdr2S40CastFromOutsideHandCostsLess(n int, label string) game.CostModifier {
	return CostsLess(n, label, YourSpell(), func(q game.CostQuery) bool {
		return q.FromZone != "" && q.FromZone != game.ZoneHand
	})
}

// cmdr2S40OpponentCastNotFromHand is "whenever an opponent casts a spell
// from anywhere other than their hand": a cast event by an opponent of the
// source's controller whose origin zone is known and is not the hand.
func cmdr2S40OpponentCastNotFromHand(ev game.Event, source *game.Card, c game.Characteristic, g *game.Game) bool {
	return ev.Kind == game.EventCast && ByAnOpponent(ev, source, c, g) &&
		ev.OldZone != "" && ev.OldZone != game.ZoneHand
}
