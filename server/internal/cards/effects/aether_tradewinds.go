package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aether Tradewinds — Instant {2}{U} (Worldwake, #5582):
//
//	"Return target permanent you control and target permanent you don't
//	 control to their owners' hands."
//
// Two target clauses (ADR 0065 §1), each judged on its own at announce
// and at resolution; the targets still legal return together, as
// Churning Eddy's do.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3a0dd318-9523-4740-b0aa-a911032e3851",
		Name:         "Aether Tradewinds",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetPermanent("target permanent you control", YouControl()),
			TargetPermanent("target permanent you don't control", OpponentControls()),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			ctx.Game.BounceCardsToHandForEffect(legalTargetCardIDs(ctx))
			return nil
		},
	})
}
