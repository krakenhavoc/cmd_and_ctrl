package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// sacrificed_this_way.go — shared bodies for the spells that read the
// permanent their additional cost sacrificed (ADR 0113 §1, #2072).
// Append-only: a new reader goes at the bottom.
//
// The record is PaidCost.SacrificedObjects, read through
// Context.SacrificedPermanents / SacrificedPower / SacrificedTotalPower:
// each sacrificed permanent as it last existed on the battlefield (CR
// 400.7j, 608.2h), so an anthem's bonus and its counters count, and a
// CR 707.10 copy reads the permanent the ORIGINAL sacrificed.

// damageEqualToSacrificedPower is "deals damage equal to the sacrificed
// creature's power to any target" (Fling, Kazuul's Fury): the first
// target takes the sacrificed creature's last-known power, floored at
// zero (CR 107.1b), so a creature with negative power deals none. A
// target that is gone or illegal by resolution is the engine's CR
// 608.2b re-check, as for every single-target damage spell.
func damageEqualToSacrificedPower(item *game.StackItem, ctx *Context) error {
	if len(item.Targets) == 0 {
		return nil
	}
	return DealDamage{
		Source: ctx.Source(),
		Target: item.Targets[0].ID,
		Amount: ctx.SacrificedPower(),
	}.Apply(ctx)
}
