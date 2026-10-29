package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// hob_a11_helpers.go — shared bodies for The Hobbit, half A, batch 11 (#4454).

// hobA11AllMilledToHand is the clause after "mill N cards, then put all <kind>
// cards from among them into your hand" (Cantankerous Keepers, Gleam of
// Death): every milled card that passes `match` and is still in the graveyard
// goes to its owner's hand, with no prompt. `o.IDs` is the list the mill's
// ThenTail was handed, the cards that really reached the graveyard (#4253); a
// card that has left it since is not taken.
func hobA11AllMilledToHand(ctx *Context, o game.TailOutcome, match CardPredicate) error {
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || !match(ctx.Game, ctx.Controller(), c) {
			continue
		}
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
