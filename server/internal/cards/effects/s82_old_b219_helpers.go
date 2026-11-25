package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b219_helpers.go — shared pieces for Rivals of Ixalan batch 219
// (S82, #5531).

// S82OldB219YouControlAnother is "you control another <subtype>" for the
// source: a permanent of the controller's other than the source with the
// subtype as it is now.
func S82OldB219YouControlAnother(ctx *Context, subtype string) bool {
	src := ctx.Source()
	return CountControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool {
		return c.InstanceID != src && c.HasSubtypeWord(subtype)
	}) > 0
}
