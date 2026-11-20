package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Double Cleave — Instant {1}{R/W} (Eventide, #5590):
//
//	"Target creature gains double strike until end of turn. (It deals both
//	 first-strike and regular combat damage.)"
//
// Ticked Off's grant.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d86f3d5e-838c-4787-9417-85a6997c3bde",
		Name:         "Double Cleave",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0, "double strike")),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return grantEachLegalTargetUntilEOT(ctx, "double strike", "Double Cleave — double strike until end of turn")
		},
	})
}
