package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "db69c3c9-d809-4361-8055-8c2dd4c6dd67",
		Name:         "Gather the White Lotus",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 2},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				n := effects.B33IslandsControlled(ctx.Game, ctx.Controller(), ctx.TextSubtype("Plains"))
				return effects.CreateToken{Template: effects.TokenCard("1/1 white Ally"), N: n}.Apply(ctx)
			}),
			effects.Scry{N: 2},
		),
	})
}
