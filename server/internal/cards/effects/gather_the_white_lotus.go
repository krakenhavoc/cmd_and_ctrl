package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gather the White Lotus — Sorcery {4}{W}:
//
//	"Create a 1/1 white Ally creature token for each Plains you
//	 control. Scry 2."
//
// The Plains are counted on resolution through effective subtypes, so
// a land given the Plains type counts (b33IslandsControlled reads any
// land subtype; "Plains" goes through the spell's text, CR 612.2).
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "db69c3c9-d809-4361-8055-8c2dd4c6dd67",
		Name:         "Gather the White Lotus",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Scry: 2},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				n := b33IslandsControlled(ctx.Game, ctx.Controller(), ctx.TextSubtype("Plains"))
				return CreateToken{Template: TokenCard("1/1 white Ally"), N: n}.Apply(ctx)
			}),
			Scry{N: 2},
		),
	})
}
