package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// DrawThenEachMayPutALand is "draw a card. Each player may put a land card
// from their hand onto the battlefield, then each opponent who didn't draws a
// card" (Anax and Cymede & Kynaios and Tiro, Kynaios and Tiro of Meletis): the
// controller draws, then each player in turn order from the controller is
// asked "may put a land card".
func DrawThenEachMayPutALand(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (DrawCards{N: 1}).Apply(ctx); err != nil {
		return err
	}
	return S82GapB21AnaxAsk(ctx, item.Controller)
}
