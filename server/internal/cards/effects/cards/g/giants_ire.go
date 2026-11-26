package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Giant's Ire — Kindred Sorcery — Giant {3}{R} (Lorwyn, S82 OLD b546,
// #5593):
//
//	"Giant's Ire deals 4 damage to target player or planeswalker. If you
//	 control a Giant, draw a card."
//
// The damage, then the draw decided as the spell resolves (CR 608.2c). A
// target that is illegal by resolution stops the whole spell, draw included
// (CR 608.2b); a changeling you control is a Giant (CR 702.73a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "73d175b3-f0dc-4599-8702-2967a8ead06a",
		Name:         "Giant's Ire",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 4)),
		Targets:      effects.TargetPlayerOrPlaneswalker(),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DamageToFirstTargetStep(4)(ctx.Item, ctx)
			}),
			effects.S82OldB546DrawIfYouControl("Giant"),
		),
	})
}
