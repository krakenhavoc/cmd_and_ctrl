package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// keyed_tails.go — #4253 slice 4: the "then" of a library search or a
// look-at-the-top (scry, surveil) written as a registered tail
// (game.RegisterTail) plus the plain data it reads, so a table waiting
// on the entry prompt of a fetched permanent, or on a scry, is a restore
// point instead of a closure that a restore point cannot carry.
//
// A card registers its tail once, from a package-level var, under an
// on-disk key (ledgered "tail <key>": never renamed, never reused), and
// passes game.KeyedTail{Ref: ref, ...} as SearchLibrary.ThenTail,
// Scry.ThenTail or Surveil.ThenTail. What the closure captured is data
// on the tail: Player and Source (the effect's controller and source
// card), N, Carry (IDs), or the whole asking item (Context.ItemTail) for
// a body that reads its targets, its X or its text.
//
// Tails that several cards share are the ones below: "then shuffle",
// "draw N cards", "gain or lose N life", "discard at random", "get N {E}"
// and "put the found lands onto the battlefield tapped".

// CardTail registers a keyed "then" whose body runs against a Context
// rebuilt from the tail: the asking item's copy when the tail carries
// one (Context.ItemTail), else a stand-in with the tail's Player as the
// controller and its Source as the source card (keyedContext). The
// outcome holds the tail's N and Carry and the IDs the event settled
// with (the cards a search found). Call it once, from a package-level
// var.
func CardTail(key string, body func(ctx *Context, o game.TailOutcome) error) game.TailRef {
	return game.RegisterTail(key, func(g *game.Game, o game.TailOutcome) error {
		return body(tailContext(g, o.Tail), o)
	})
}

// tailContext is the Context a card's keyed tail runs against.
func tailContext(g *game.Game, t game.KeyedTail) *Context {
	if item := t.Asked.StackItem(); item != nil {
		return NewContext(g, item)
	}
	return keyedContext(g, t.Player, t.Source)
}

// Tail is a keyed "then" for ref with this effect's controller and
// source card filled in: set Player, N or Carry on the result when the
// body reads them.
func (c *Context) Tail(ref game.TailRef) game.KeyedTail {
	return game.KeyedTail{Ref: ref, Player: c.Controller(), Source: c.Source()}
}

// TailN is Tail with the number the body reads.
func (c *Context) TailN(ref game.TailRef, n int) game.KeyedTail {
	t := c.Tail(ref)
	t.N = n
	return t
}

// ItemTail is Tail with the resolving item carried as data too, for a body
// that reads more of it than its controller and source (its targets, its X,
// the words of its text, the window its source object had).
func (c *Context) ItemTail(ref game.TailRef) game.KeyedTail {
	t := c.Tail(ref)
	t.Asked = game.NewAskedItem(c.Item)
	return t
}

// thenShuffle is "then shuffle", Tail.Player's library. An on-disk key.
var thenShuffle = game.RegisterTail("then/shuffle", func(g *game.Game, o game.TailOutcome) error {
	return g.ShuffleLibraryForEffect(o.Tail.Player)
})

// thenDrawCards is "then draw N cards", Tail.N of them, for Tail.Player.
// An on-disk key.
var thenDrawCards = game.RegisterTail("then/draw-n-cards", func(g *game.Game, o game.TailOutcome) error {
	return g.DrawNForEffect(o.Tail.Player, o.Tail.N)
})

// thenChangeLife is "you gain N life" (N > 0) or "you lose N life"
// (N < 0), Tail.Player's life, attributed to Tail.Source. An on-disk key.
var thenChangeLife = game.RegisterTail("then/change-life", func(g *game.Game, o game.TailOutcome) error {
	return g.ChangePlayerLifeForEffect(o.Tail.Source, o.Tail.Player, o.Tail.N)
})

// thenDiscardAtRandom is "then discard N cards at random", Tail.N of
// Tail.Player's. An on-disk key.
var thenDiscardAtRandom = game.RegisterTail("then/discard-at-random", func(g *game.Game, o game.TailOutcome) error {
	return g.DiscardRandomForEffect(o.Tail.Player, o.Tail.N)
})

// thenGetEnergy is "you get N {E}", Tail.Player's. An on-disk key.
var thenGetEnergy = game.RegisterTail("then/get-energy", func(g *game.Game, o game.TailOutcome) error {
	return g.AddPlayerCounterByForEffect(o.Tail.Player, o.Tail.Player, game.CounterEnergy, o.Tail.N)
})

// thenPutFoundLandsTapped is "if it's a land card, put it onto the
// battlefield tapped" over the cards a search put into the hand: each
// land among them (those still in the hand) enters under Tail.Player's
// control. An on-disk key.
var thenPutFoundLandsTapped = game.RegisterTail("then/put-found-lands-onto-battlefield-tapped",
	func(g *game.Game, o game.TailOutcome) error {
		for _, id := range o.IDs {
			c, ok := g.LookupCardForEffect(id)
			if !ok || !c.IsLand() {
				continue
			}
			if _, err := g.PutFromHandOntoBattlefieldForEffect(id, game.HandEntryOptions{Controller: o.Tail.Player, Tapped: true}); err != nil {
				return err
			}
		}
		return nil
	})
