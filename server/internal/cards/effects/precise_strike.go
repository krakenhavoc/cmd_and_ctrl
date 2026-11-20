package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Precise Strike — Instant {R} (Aether Revolt,
// #5536):
//
//	"Target creature gets +1/+0 and gains first strike until end of turn."
//
// Kindled Fury.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bfd2138b-1883-4900-a5f1-373884965c34",
		Name:         "Precise Strike",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 1, 0, "first strike")),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return pumpAndGrantFirstTarget(1, 0, "Precise Strike — +1/+0 and first strike", "first strike")(ctx.Game, item)
		},
	})
}
