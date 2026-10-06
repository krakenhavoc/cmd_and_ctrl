package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Coursers' Accord — Sorcery {4}{G}{W}:
//
//	"Create a 3/3 green Centaur creature token, then populate.
//	 (Create a token that's a copy of a creature token you control.)"
//
// The token is made first, so it is a candidate for the populate that
// follows (and the only one when you control no other creature token).
// Populate is the shared primitive in populate.go.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "23674648-d1a0-437d-8a0b-4173c75b4507",
		Name:         "Coursers' Accord",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (CreateToken{Template: TokenCard("3/3 green Centaur"), N: 1}).Apply(ctx); err != nil {
				return err
			}
			return Populate{}.Apply(ctx)
		},
	})
}
