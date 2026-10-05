package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Horncaller's Chant — Sorcery {7}{G}:
//
//	"Create a 4/4 green Rhino creature token with trample, then
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
		OracleID:     "8c069834-ed89-4298-989b-e67036d56196",
		Name:         "Horncaller's Chant",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (CreateToken{Template: TokenCard("4/4 green Rhino with trample"), N: 1}).Apply(ctx); err != nil {
				return err
			}
			return Populate{}.Apply(ctx)
		},
	})
}
