package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kindled Fury — Instant {R} (Foundations, #4465):
//
//	"Target creature gets +1/+0 and gains first strike until end of
//	 turn."
//
// Kindled Heroism without the scry: pumpAndGrantFirstTarget on the target
// still legal as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ebeee06e-3345-4f02-896f-cb0b2cfa5548",
		Name:         "Kindled Fury",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 1, 0, "first strike")),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return pumpAndGrantFirstTarget(1, 0, "Kindled Fury — +1/+0 and first strike", "first strike")(ctx.Game, item)
		},
	})
}
