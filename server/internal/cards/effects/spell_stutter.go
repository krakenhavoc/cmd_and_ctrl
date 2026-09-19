package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spell Stutter — Instant {1}{U} (EDHREC rank 4022):
//
//	"Counter target spell unless its controller pays {2} plus an
//	 additional {1} for each Faerie you control."
//
// Mana Leak that scales with the board. In a Faerie deck with three
// bodies out it is "unless its controller pays {5}", which at instant
// speed on an opponent's turn is usually uncounterable in practice.
//
// It is in the batch because the tax is COMPUTED AT RESOLUTION from
// the board rather than printed, which is a shape no other
// counterspell in the catalog has — Daze, Mana Leak and Spell Pierce
// all name a constant.
//
// # When the Faeries are counted
//
// On resolution, not on announce. A Faerie that died in response
// lowers the tax; a Faerie flashed in raises it. That is CR 608.2's
// ordinary reading — the number is not locked in when the spell is
// cast — and it is the reason the cost string is built here rather
// than declared on the Spec.
//
// "You control" is the Stutter's CASTER, not the countered spell's
// controller, and it is read post-layer, so a changeling counts as a
// Faerie and a Faerie that lost the type does not. The Stutter itself
// is not a Faerie and never counts.
//
// # It is a pay-unless, not a counter
//
// The victim gets the choice, which is why the tax matters at all:
// the spell resolves if they pay and is countered if they do not or
// cannot. A "can't be countered" spell is still a legal target — the
// prompt is still offered and the counter half simply does nothing
// (CR 701.6a), which is observably different from the Stutter
// fizzling.
//
// The payer is read off the stack BEFORE anything touches it, because
// countering deletes the stack entry the prompt would otherwise have
// to look the payer up from.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f9bf996e-ae14-40a3-a8d8-f5725ddaa270",
		Name:         "Spell Stutter",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			stackID := item.Targets[0].ID
			target := ctx.Game.StackItemForEffect(stackID)
			if target == nil {
				return nil
			}
			tax := 2 + b38FaeriesYouControl(ctx.Game, item.Controller)
			cost := "{" + strconv.Itoa(tax) + "}"
			return CounterUnlessPaid{
				StackID:  stackID,
				Cost:     cost,
				Question: "Spell Stutter — pay " + cost + " or your spell is countered",
			}.Apply(ctx)
		},
	})
}
