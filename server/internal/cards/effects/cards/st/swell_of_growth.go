package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Swell of Growth — Instant {1}{G} (Battle for Zendikar):
//
//	"Target creature gets +2/+2 until end of turn. You may put a land card
//	 from your hand onto the battlefield."
//
// The land is offered even when the creature is no longer a legal target only
// if the spell still resolves (CR 608.2b: a spell with one illegal target does
// nothing).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e47e0b81-0f0c-4e52-bc58-16af19b0dfe0",
		Name:         "Swell of Growth",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      swellOfGrowthPurpose(),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.BoostTargetUntilEOT("Swell of Growth — +2/+2 until end of turn", 2, 2)(ctx.Game, ctx.Item)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.MayPutALandFromHand("Swell of Growth").Apply(ctx)
			}),
		),
	})
}

// swellOfGrowthPurpose declares the +2/+2 on its target and the land it may
// put onto the battlefield.
func swellOfGrowthPurpose() game.Purpose {
	p := effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2))
	p.Lands = 1
	return p
}
