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
// The player chooses what to proliferate (proliferate.go, #2525), and
// the draw is the proliferate's Then: it must come AFTER the answer,
// or the card is drawn while the prompt is still open.
func init() {
	Register(Spec{
		OracleID: "d2145ad3-fe7e-459b-b155-1b3ed089e936",
		Name:     "Steady Progress",
		// Every printed clause happens: the choice is the player's
		// and the draw follows it.
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			controller := ctx.Controller()
			return Proliferate{Then: func(g *game.Game) error {
				return g.DrawNForEffect(controller, 1)
			}}.Apply(ctx)
		},
	})
}
