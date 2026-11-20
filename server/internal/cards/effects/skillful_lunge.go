package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skillful Lunge — Instant {1}{W} (Dark Ascension, #5571):
//
//	"Target creature gets +2/+0 and gains first strike until end of
//	 turn."
//
// Kindled Fury with +2/+0: pumpAndGrantFirstTarget on the target still
// legal as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bc730e60-b9c2-4c49-b97b-2878be105b04",
		Name:         "Skillful Lunge",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 0, "first strike")),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return pumpAndGrantFirstTarget(2, 0, "Skillful Lunge — +2/+0 and first strike", "first strike")(ctx.Game, item)
		},
	})
}
