package effects

import "github.com/google/uuid"

// msc5_shim.go — two bodies that PR #4582 (MSC batch 3) also defines, under
// the same names. This file exists only until that PR lands; delete it on the
// rebase that brings them in.

// sourceConnives makes the ability's source connive (CR 701.50).
func sourceConnives(ctx *Context) error {
	return ConniveEach{Controller: ctx.Controller(), Creatures: []uuid.UUID{ctx.Item.SourceCardID}}.Apply(ctx)
}

// boomerangPingAndGain is one damage to every opponent, then one life for the
// controller.
func boomerangPingAndGain(ctx *Context) error {
	if err := damageToEachOpponent(ctx.Game, ctx.Item, 1); err != nil {
		return err
	}
	return GainLife{Player: ctx.Controller(), Amount: 1}.Apply(ctx)
}
