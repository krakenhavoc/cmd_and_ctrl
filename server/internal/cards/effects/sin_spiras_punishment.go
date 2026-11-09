package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sin, Spira's Punishment — Legendary Creature — Leviathan Avatar
// {4}{B}{G}{U}, 7/7 (Final Fantasy, #4462):
//
//	"Flying
//	 Whenever Sin enters or attacks, exile a permanent card from your
//	 graveyard at random, then create a tapped token that's a copy of that
//	 card. If the exiled card is a land card, repeat this process."
//
// One pass is: choose a permanent card (CR 110.1: not an instant or a
// sorcery) from the controller's graveyard at random, exile it, and create
// a tapped token copy of it (CR 707.2, as the card is in exile; an {X} in
// its cost counts as 0). Each step waits for the one before as a keyed
// tail, so a pass that stops to ask something (a commander's zone, a copied
// Clone) finishes before the next begins. The land question is read from the
// card before it leaves the graveyard, and only a card that did reach exile
// repeats the process; an empty graveyard, or one with no permanent card,
// ends it. A land copy enters tapped like any other.
//
// No simplification.
func init() {
	sinRepeat = sinPass
	Register(Spec{
		OracleID:        "3bcc378c-4470-4757-a0d5-025a32c918ea",
		Name:            "Sin, Spira's Punishment",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Sin, Spira's Punishment — exile a permanent card from your graveyard at random, then create a tapped token that's a copy of it; repeat if it was a land",
				func(g *game.Game, item *game.StackItem) error { return sinPass(NewContext(g, item)) }),
		},
	})
}

// sinPass is one pass of the process.
func sinPass(ctx *Context) error {
	ids := graveyardIDs(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.IsPermanent() })
	pick := ctx.Game.ChooseAtRandomForEffect(randomDraw(ctx), ids, 1)
	if len(pick) == 0 {
		return nil
	}
	isLand := 0
	if c, ok := ctx.Game.LookupCardForEffect(pick[0]); ok && c.IsLand() {
		isLand = 1
	}
	return ctx.Game.ExileCardKeyedForEffect(pick[0], withN(ctx.ItemTail(sinAfterExile), isLand))
}

// sinAfterExile creates the tapped copy of the card that reached exile. An
// on-disk key.
var sinAfterExile = CardTail("sin-spiras-punishment/copy-exiled", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return finB9CreateTappedCopy(ctx, o.IDs[0], withN(ctx.ItemTail(sinAfterCopy), o.Tail.N))
})

// sinAfterCopy repeats the process when the exiled card was a land card. An
// on-disk key.
var sinAfterCopy = CardTail("sin-spiras-punishment/repeat-for-a-land", func(ctx *Context, o game.TailOutcome) error {
	if o.Tail.N != 1 {
		return nil
	}
	return sinRepeat(ctx)
})

// sinRepeat is sinPass, reached through a variable set in init so the
// tails and the pass that queues them are not an initialisation cycle.
var sinRepeat func(*Context) error
