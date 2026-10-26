package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// counter_tails.go — #4253 slice 7: the "then" of a counter placement and of
// a life change written as a registered tail (game.RegisterTail, CardTail),
// so a table waiting on the CR 616 ordering prompt over the placement (a
// Doubling Season beside a Hardened Scales) is a restore point instead of a
// closure a restore point cannot carry.
//
// A card passes one of these as the `then` of AddCounterByKeyedForEffect or
// ChangePlayerLifeKeyedForEffect, built with ctx.ItemTail (the asking item
// rides it as data) or ctx.Tail. What the closure captured is data on the
// tail: Carry (the permanent, or the ones still owed), N and Word. The tails
// several cards share are the ones below; the rest sit beside their card.

// thenDoublePlusOneCounters is "double the number of +1/+1 counters on it"
// once the first counter has landed (Invigorating Surge): Carry[0] is the
// creature, and it is doubled only if it is still on the battlefield with
// some. An on-disk key: never renamed, never reused.
var thenDoublePlusOneCounters = CardTail("counter/double-plus-one", func(ctx *Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	return b23DoublePlusOneCountersOn(ctx, o.Tail.Carry[0])
})

// thenDrawPerCounter is "draw a card for each <kind> counter on it" once the
// counter has landed (Insight Engine's charge, The One Ring's burden):
// Carry[0] is the permanent, Word the counter kind, and Player draws. An
// on-disk key: never renamed, never reused.
var thenDrawPerCounter = CardTail("counter/draw-per-counter", func(ctx *Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(o.Tail.Carry[0])
	if !ok {
		return nil
	}
	return DrawCards{Player: o.Tail.Player, N: c.Counters[o.Tail.Word]}.Apply(ctx)
})

// thenGainLifeEqualToLost is "you gain life equal to the life lost this way"
// (Devouring Greed): the tail's Player gains what the loss reported, and the
// gain is attributed to its Source. A change that did not lose (the loss was
// replaced away, or became a gain) gains nothing. An on-disk key: never
// renamed, never reused.
var thenGainLifeEqualToLost = game.RegisterTail("life/gain-equal-to-lost", func(g *game.Game, o game.TailOutcome) error {
	if o.Amount >= 0 {
		return nil
	}
	return g.ChangePlayerLifeForEffect(o.Tail.Source, o.Tail.Player, -o.Amount)
})

// carryFirst is the first of a tail's Carry, or uuid.Nil.
func carryFirst(t game.KeyedTail) uuid.UUID {
	if len(t.Carry) == 0 {
		return uuid.Nil
	}
	return t.Carry[0]
}
