package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Steady Progress — Instant {2}{U}:
//
//	"Proliferate. (Choose any number of permanents and/or players,
//	 then give each another counter of each kind already there.)
//	 Draw a card."
//
// The plainest possible proliferate card, and the first one in the
// catalog: no trigger, no cost, no condition — just the keyword
// action and a cantrip. It exists to pin the primitive end to end.
//
// The choice is made by the auto-pick in proliferate.go rather than
// prompted; see that file for what the pick takes and where it
// differs from paper.
func init() {
	Register(Spec{
		OracleID: "d2145ad3-fe7e-459b-b155-1b3ed089e936",
		Name:     "Steady Progress",
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (Proliferate{}).Apply(ctx); err != nil {
				return err
			}
			return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
