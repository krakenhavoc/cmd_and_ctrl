package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b25_helpers.go — the shared pieces of Standard-set batch 25
// (Duskmourn leftovers, #4452, #4466).

// b25CondGraveyardTypes is the intervening "if there are N or more card
// types among cards in your graveyard" (delirium, CR 207.2c, and the
// six-type form Omnivorous Flytrap prints): read as the ability resolves
// (CR 603.4), through b16CardTypesInGraveyard, the count every delirium
// card reads. It declares no footprint, so a row that carries it keeps its
// trigger-order prompt.
type b25CondGraveyardTypes struct{ AtLeast int }

func (c b25CondGraveyardTypes) Holds(ctx *Context) bool {
	return b16CardTypesInGraveyard(ctx.Game, ctx.Controller()) >= c.AtLeast
}

func (c b25CondGraveyardTypes) condition() (game.FootprintStep, bool) {
	return game.FootprintStep{}, false
}

// b25YourGraveyardTypesAtLeast is the trigger-time half of that "if": the
// When a row's AppliesTo asks as the event is seen (CR 603.4).
func b25YourGraveyardTypesAtLeast(n int) When {
	return func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return b16CardTypesInGraveyard(g, source.Controller) >= n
	}
}
