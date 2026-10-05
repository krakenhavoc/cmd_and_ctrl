package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eyes in the Skies — Instant {3}{W}:
//
//	"Create a 1/1 white Bird creature token with flying, then
//	 populate. (Create a token that's a copy of a creature token you
//	 control.)"
//
// The token is made first, so it is a candidate for the populate that
// follows (and the only one when you control no other creature token).
// Populate is the shared primitive in populate.go.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9697cd05-474e-46f3-8bcc-d4b6fb2059a1",
		Name:         "Eyes in the Skies",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (CreateToken{Template: TokenCard("1/1 white Bird with flying"), N: 1}).Apply(ctx); err != nil {
				return err
			}
			return Populate{}.Apply(ctx)
		},
	})
}
