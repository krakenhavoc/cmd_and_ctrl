package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reckless Spite — Instant {1}{B}{B}:
//
//	"Destroy two target nonblack creatures. You lose 5 life."
//
// Curtains' Call's destruction: one simultaneous event over the targets
// still legal as the spell resolves (CR 608.2b; a death trigger among them
// sees both). "Nonblack" is judged at announce and again at resolution. The
// life is lost whenever the spell resolves at all, so with one target gone
// it still costs 5; with both gone it fizzles and costs nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a684df3a-5441-4daa-86d1-c47a91b35e6a",
		Name:         "Reckless Spite",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("two target nonblack creatures", effects.NonBlack()).WithCount(2, 2),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.DestroyEachLegalTarget(ctx.Item, ctx) }),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.GainLife{Player: ctx.Controller(), Amount: -5}.Apply(ctx)
			}),
		),
	})
}
