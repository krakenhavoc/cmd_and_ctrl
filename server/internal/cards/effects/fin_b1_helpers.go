package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// fin_b1_helpers.go — shared helpers for Final Fantasy, batch 1 (#4462).

// finB1CondEightInYourGraveyard is the intervening "if eight or more
// cards are in your graveyard" (Sidequest: Card Collection), re-checked
// as the ability resolves (CR 603.4). It reads the item's controller's
// graveyard (game.FootprintCondGraveyard).
type finB1CondEightInYourGraveyard struct{}

func (finB1CondEightInYourGraveyard) Holds(ctx *Context) bool {
	return GraveyardAtLeast(8, nil)(ctx.Game, ctx.Controller(), ctx.Source())
}

func (finB1CondEightInYourGraveyard) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondGraveyard, 0), true
}
