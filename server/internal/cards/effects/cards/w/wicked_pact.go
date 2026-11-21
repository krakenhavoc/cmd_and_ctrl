package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wicked Pact — Sorcery {1}{B}{B} (Portal, #5630):
//
//	"Destroy two target nonblack creatures. You lose 5 life."
//
// Reckless Spite's text, word for word. The life is lost even if one
// target has become illegal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "39e21a5a-b278-478a-854c-17695c0f6246",
		Name:         "Wicked Pact",
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
