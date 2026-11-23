package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// DrawACardAndLoseOneLifeMode is the bullet body "You draw a card and you
// lose 1 life." for a mode with no target (Disciple of Perdition, Dread
// Presence): the controller draws, then loses 1 life, in printed order
// (CR 608.2c).
func DrawACardAndLoseOneLifeMode(item *game.StackItem, ctx *Context, _ int) error {
	if err := (DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
		return err
	}
	return ctx.Game.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -1)
}
