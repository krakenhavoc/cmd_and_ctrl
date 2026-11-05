package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ticked Off — {1}{R} Sorcery (Teenage Mutant Ninja Turtles Commander,
// #5011):
//
//	"Target creature gains double strike until end of turn."
//
// A layer-6 grant pinned to the target until the cleanup step (CR 514.2).
// A target that left or stopped being a creature gets nothing, and with no
// legal target the spell does not resolve (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "91b89c21-5f35-4aa5-925b-c756ec6969b4",
		Name:         "Ticked Off",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0, "double strike")),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return grantEachLegalTargetUntilEOT(ctx, "double strike", "Ticked Off — double strike until end of turn")
		},
	})
}
