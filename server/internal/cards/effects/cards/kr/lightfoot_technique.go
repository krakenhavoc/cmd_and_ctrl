package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lightfoot Technique — Instant {1}{W} (Tarkir: Dragonstorm, #4463):
//
//	"Put a +1/+1 counter on target creature. It gains flying and
//	 indestructible until end of turn."
//
// Fleeting Flight's first two sentences with indestructible beside the
// flying: the counter stays, and the keywords end in the cleanup step (CR
// 514.2). The keywords wait for a prompt the counter asks (CR 608.2c), and
// a target that has left by resolution gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "601b7481-19fc-4779-950e-883a9a22aec3",
		Name:         "Lightfoot Technique",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.PlusOneCounterOnTheTarget(ctx.Game, ctx.Item) }),
			effects.Instruction(func(ctx *effects.Context) error {
				if id, ok := effects.B16FirstLegalTargetCard(ctx); ok {
					return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying", "indestructible"},
						Label: "Lightfoot Technique — flying and indestructible"}.Apply(ctx)
				}
				return nil
			}),
		),
	})
}
