package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Epic is a spell's resolution with its epic ability (CR 702.50a, #6141):
//
//	"Epic (For the rest of the game, you can't cast spells. At the
//	 beginning of each of your upkeeps, copy this spell except for its
//	 epic ability. If the spell has any targets, you may choose new
//	 targets for the copy.)"
//
// The two epic abilities are applied first, while the spell is certainly
// still the resolving object, and then the spell's own text (body). The
// order is not observable: neither epic ability does anything during this
// resolution. A copy the epic trigger makes has lost its epic ability
// (CR 707.9b), so its resolution runs body alone. See game/epic.go.
func Epic(body func(item *game.StackItem, ctx *Context) error) func(item *game.StackItem, ctx *Context) error {
	return func(item *game.StackItem, ctx *Context) error {
		ctx.Game.EpicForEffect(item.SourceCardID, "")
		return body(item, ctx)
	}
}
