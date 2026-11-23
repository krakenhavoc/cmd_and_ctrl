package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lofty Denial — Instant {1}{U} (Jumpstart Azorius deck request, #4894):
//
//	"Counter target spell unless its controller pays {1}. If you
//	 control a creature with flying, counter that spell unless its
//	 controller pays {4} instead."
//
// Dazzling Denial's shape. The flyer is looked for at RESOLUTION, part
// of the effect rather than a condition on casting it, so a flyer that
// arrives in response raises the tax to {4} and one killed in response
// lowers it to {1}. "A creature with flying" is read after the layers:
// a creature granted flying counts, one that lost it does not. The
// prompt goes to the spell's controller (CounterUnlessPaid reads the
// payer off the guarded stack item), and a payment they cannot fund
// counters the spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d799a2c4-628a-45da-b9ca-0337c46e3e14",
		Name:         "Lofty Denial",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			stackID := item.Targets[0].ID
			if ctx.Game.StackItemForEffect(stackID) == nil {
				return nil
			}
			cost := "{1}"
			if effects.B40ControlsA(ctx.Game, ctx.Controller(), effects.And(effects.Creature(), effects.HasKeyword("flying"))) {
				cost = "{4}"
			}
			return effects.CounterUnlessPaid{
				StackID:  stackID,
				Cost:     cost,
				Question: "Lofty Denial — pay " + cost + " or your spell is countered",
			}.Apply(ctx)
		},
	})
}
