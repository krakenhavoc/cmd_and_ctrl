package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OldB227RowdyCrewETB is Rowdy Crew's enters trigger: "draw three cards,
// then discard two cards at random. If two cards that share a card type are
// discarded this way, put two +1/+1 counters on this creature."
//
// The draw happens first and settles; the random discard then runs as one
// batch, and OldB227RowdyCrewCounters, told the cards really discarded
// (CR 701.9a), checks them. Fewer than two discarded (a discard another
// effect stopped) is not "two cards", so no counters.
func OldB227RowdyCrewETB(g *game.Game, item *game.StackItem) error {
	if err := g.DrawNForEffect(item.Controller, 3); err != nil {
		return err
	}
	return g.DiscardRandomKeyedForEffect(item.Controller, 2, game.KeyedTail{
		Ref:    OldB227RowdyCrewCounters,
		Player: item.Controller,
		Source: item.SourceCardID,
		Asked:  game.NewAskedItem(item),
	})
}

// OldB227RowdyCrewCounters is the rest of Rowdy Crew's trigger, run once the
// random discard has landed: two +1/+1 counters on the Crew if exactly two
// cards were discarded and some card type is common to both. The Crew is
// "this creature": a Crew that left the battlefield, or came back as a new
// object, gets nothing (CR 400.7). An on-disk key.
var OldB227RowdyCrewCounters = game.RegisterTail("tail/rowdy-crew-counters", func(g *game.Game, o game.TailOutcome) error {
	if len(o.IDs) != 2 {
		return nil
	}
	var cards []*game.Card
	for _, id := range o.IDs {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			return nil
		}
		cards = append(cards, &c)
	}
	if !trekB23ShareACardType(cards) {
		return nil
	}
	item := o.Tail.Asked.StackItem()
	if item == nil {
		return nil
	}
	return AddCounter{Target: o.Tail.Source, Kind: game.CounterPlusOne, N: 2}.Apply(NewContext(g, item))
})
