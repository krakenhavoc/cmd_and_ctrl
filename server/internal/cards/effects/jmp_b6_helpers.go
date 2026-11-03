package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// jmp_b6_helpers.go — bodies shared by two cards of the Jumpstart deck
// request batch 6 (#4893). Append-only.

// mayFetchABasicLandToHand is the enters trigger body "you may search
// your library for a basic land card, reveal it, put it into your
// hand, then shuffle" (Sylvan Ranger, Skittering Surveyor). Optional is
// the printed "you may": declining skips the search and the shuffle.
// `name` names the card on the search prompt.
func mayFetchABasicLandToHand(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return SearchLibrary{
			Player:    item.Controller,
			Predicate: IsBasicLand,
			Dest:      game.ZoneHand,
			Limit:     1,
			Reveal:    true,
			Shuffle:   true,
			Optional:  true,
			Reason:    name + " — a basic land card",
		}.Apply(NewContext(g, item))
	}
}

// putAPlusOneCounterOnThisCreatureMode is the modal bullet "Put a +1/+1
// counter on this creature." (Trufflesnout, Trusty Retriever). A
// source that has left the battlefield is a new object, or no object,
// so nothing gets the counter (CR 400.7).
func putAPlusOneCounterOnThisCreatureMode(item *game.StackItem, ctx *Context, _ int) error {
	if !onBattlefield(ctx.Game, item.SourceCardID) {
		return nil
	}
	return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
}
