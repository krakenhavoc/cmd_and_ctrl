package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mass_route_tails.go — #4253 slice 10a: the "then" of a mass verb
// (DestroyAllMatching, ExileAllMatching, BounceAllMatching) and of a
// single sacrifice, bounce, destroy or tuck, written as a registered tail
// so a table waiting on a CR 616 prompt over one of the batch's legs is a
// restore point. A mass verb's tail is registered with MassTail and told
// the pre-move copies of what went this way; a single-card door's tail is
// told the card when it went this way and nothing when it did not
// (oneWentThisWay). Keys are on-disk identities (ledgered "tail <key>"):
// never renamed, never reused.
//
// The tails below are the ones several cards share.

// massGainLifeForEach is "you gain 1 life for each <permanent> <verb>
// this way" (Fumigate, Avenge, Ob Nixilis, the Ascended).
var massGainLifeForEach = MassTail("mass/gain-life-for-each", func(ctx *Context, _ []game.Card, moved int) error {
	if moved <= 0 {
		return nil
	}
	return GainLife{Player: ctx.Controller(), Amount: moved}.Apply(ctx)
})

// thenVentureIntoTheDungeon is "then venture into the dungeon"
// (Acererak, the Archlich's return).
var thenVentureIntoTheDungeon = CardTail("then/venture-into-the-dungeon", func(ctx *Context, _ game.TailOutcome) error {
	return VentureIntoTheDungeon{}.Apply(ctx)
})

// thenManifestDread is "then manifest dread" after a bounce (Glitch
// Interpreter, Unnerving Grasp).
var thenManifestDread = CardTail("then/manifest-dread", func(ctx *Context, _ game.TailOutcome) error {
	return ManifestDread{}.Apply(ctx)
})

// thenEachOpponentLosesN is "each opponent loses N life", Tail.N of it,
// attributed to the source (Sorin, Grim Nemesis; Caustic Bronco saddled).
// The asking item rides the tail (ItemTail).
var thenEachOpponentLosesN = CardTail("then/each-opponent-loses-n-life", func(ctx *Context, o game.TailOutcome) error {
	if o.Tail.N <= 0 {
		return nil
	}
	return eachOpponentLosesLife(ctx.Game, ctx.Item, o.Tail.N)
})

// thenIfItWentRun is "if you do, <steps>": the steps registered under the
// tail's Word (RegisterThenSteps) run against the asking item when the
// one card went this way (oneWentThisWay).
var thenIfItWentRun = CardTail("then/if-it-went-run-steps", func(ctx *Context, o game.TailOutcome) error {
	if !oneWentThisWay(o) {
		return nil
	}
	steps, ok := thenStepsByKey[o.Tail.Word]
	if !ok {
		return nil
	}
	return ThenSteps{key: o.Tail.Word, steps: steps}.Run(ctx)
})

// ifItWentThen is thenIfItWentRun over steps, for the asking item.
func (c *Context) ifItWentThen(steps ThenSteps) game.KeyedTail {
	return withWord(c.ItemTail(thenIfItWentRun), steps.Key())
}

// thenShuffleEachCarried is "then each of those owners shuffles", in the
// order the caller put them on the tail's Carry (Serene Remembrance,
// Teferi, Temporal Pilgrim).
var thenShuffleEachCarried = game.RegisterTail("then/shuffle-each-carried", func(g *game.Game, o game.TailOutcome) error {
	for _, p := range o.Tail.Carry {
		if err := g.ShuffleLibraryForEffect(p); err != nil {
			return err
		}
	}
	return nil
})

// tuckEachOwnerOrders is "each owner puts the cards they own among them
// on the <bottom | top or bottom> of their library in any order" after a
// tuck (Terminus, Trial of a Time Lord, Aetherspouts), told the cards that
// reached a library: N is the placement (tuckOrderBottom or
// tuckOrderTopOrBottom), Word the prompt's label. An owner with one card
// to put on the bottom has nothing to order.
var tuckEachOwnerOrders = CardTail("tuck/each-owner-orders", func(ctx *Context, o game.TailOutcome) error {
	place, least := game.LibraryPlaceBottom, 2
	if o.Tail.N == tuckOrderTopOrBottom {
		place, least = game.LibraryPlaceTopOrBottom, 1
	}
	byOwner := map[uuid.UUID][]uuid.UUID{}
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok {
			byOwner[c.Owner] = append(byOwner[c.Owner], id)
		}
	}
	for _, p := range ctx.Game.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		own := byOwner[p.ID]
		if len(own) < least {
			continue
		}
		if err := (PutInLibraryInAnyOrder{
			Chooser:   p.ID,
			Cards:     own,
			From:      game.ZoneLibrary,
			Placement: place,
			Label:     o.Tail.Word,
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})

// The placements tuckEachOwnerOrders reads from its N.
const (
	tuckOrderBottom      = 0
	tuckOrderTopOrBottom = 1
)

// thenShuffleIfItWent is "if you do, its owner shuffles": Tail.Player's
// library, when the one card went this way (Lembas).
var thenShuffleIfItWent = game.RegisterTail("then/shuffle-if-it-went", func(g *game.Game, o game.TailOutcome) error {
	if !oneWentThisWay(o) {
		return nil
	}
	return g.ShuffleLibraryForEffect(o.Tail.Player)
})
