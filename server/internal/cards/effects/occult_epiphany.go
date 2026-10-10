package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Occult Epiphany — Instant {X}{U}:
//
//	"Draw X cards, then discard X cards. Create a 1/1 white Spirit
//	 creature token with flying for each card type among cards
//	 discarded this way."
//
// The discard waits for the draws (CR 608.2c) so a drawn card is a
// legal discard; the tokens are made from the discard's continuation,
// once the cards are in the graveyard. A card type is counted once
// however many discarded cards share it, and a hand smaller than X
// discards what it has.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6df1c314-b97b-4bbc-8b7e-a07785347a49",
		Name:         "Occult Epiphany",
		Completeness: CompletenessFull,
		XMatters:     true,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x := ctx.X()
			if x <= 0 {
				return nil
			}
			return ctx.Game.DrawNThenForEffect(item.Controller, x, game.DrawThen{
				Ref: rfReprintAOccultEpiphanyDiscard, Player: item.Controller, Source: item.SourceCardID, N: x,
			})
		},
	})
}
