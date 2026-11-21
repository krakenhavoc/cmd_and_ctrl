package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leo's Guidance — Instant {2}{W} (Teenage Mutant Ninja Turtles
// Commander, #5011):
//
//	"Put a +1/+1 counter on each of up to three target creatures. Untap
//	 them."
//
// Up to three creature targets, zero a legal choice. Each target still
// legal as the spell resolves (CR 608.2b) gets a counter, and then those
// same targets are untapped: the untap waits for any prompt the counters
// ask (a CR 616 order).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f386bca6-a08e-462b-b1d9-c3b03929c68d",
		Name:         "Leo's Guidance",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("up to three target creatures").WithCount(0, 3),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.PutPlusOneCounterOnEachLegalTarget(ctx.Game, ctx.Item)
			}),
			effects.Instruction(func(ctx *effects.Context) error { return effects.UntapEachLegalTarget(ctx.Game, ctx.Item) }),
		),
	})
}
