package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spiritual Visit — Instant — Arcane {W}:
//
//	"Create a 1/1 colorless Spirit creature token.
//	Splice onto Arcane {W}"
func init() {
	Register(Spec{
		OracleID:     "40759025-91e1-4755-a8cf-0b68c8e9220f",
		Name:         "Spiritual Visit",
		Completeness: CompletenessFull,
		Splice:       SpliceOntoArcane("{W}"),
		Purpose:      game.Purpose{Tokens: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Controller: ctx.Controller(), Template: TokenCard("1/1 colorless Spirit"), N: 1}.Apply(ctx)
		},
	})
}
