package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// std_b27_helpers.go — shared pieces of the Standard leftovers redo set
// (feat/std-redo-27).

// CondYourGraveyardAtLeast is "if there are N or more cards in your
// graveyard" (threshold, N = 7) declared as an intervening if
// (InterveningIf): asked as the ability triggers and again as it
// resolves (CR 603.4). The graveyard is the resolving item's
// controller's, which for an ability that works from a graveyard is the
// card's owner (CR 108.4a).
type CondYourGraveyardAtLeast struct{ N int }

func (c CondYourGraveyardAtLeast) Holds(ctx *Context) bool {
	return b31GraveyardSize(ctx.Game, ctx.Controller()) >= c.N
}

func (CondYourGraveyardAtLeast) condition() (game.FootprintStep, bool) {
	return game.FootprintStep{}, false
}
