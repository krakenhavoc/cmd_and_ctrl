package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Get a Leg Up — Instant {G} (Murders at Karlov Manor, #4469):
//
//	"Until end of turn, target creature gets +1/+1 for each creature you
//	 control and gains reach."
//
// The count is Friendly Neighborhood's: the creatures the caster controls
// as the spell resolves, counted once (CR 608.2h), so the bonus does not
// change if that number does later in the turn. The target counts itself
// when the caster controls it. A target that left in response gets
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a80519c4-f778-4b8e-9920-64013e98c95b",
		Name:         "Get a Leg Up",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "reach")),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			n := effects.CountControlled(ctx.Game, item.Controller, effects.MatchCreature)
			return effects.PumpAndGrantFirstTarget(n, n, "Get a Leg Up — +1/+1 for each creature you control and reach", "reach")(ctx.Game, item)
		},
	})
}
