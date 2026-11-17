package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phyrexian Espionage — {2}{U} Sorcery (Dominaria United):
//
//	"Kicker {1}{B} (You may pay an additional {1}{B} as you cast this
//	 spell.)
//	 Draw two cards. If this spell was kicked, each opponent discards a
//	 card."
//
// Two steps (ResolveSteps): the discards are asked after the draw.
func init() {
	Register(Spec{
		OracleID:      "32dce35e-182d-41be-9fb2-b0738d0781b8",
		Name:          "Phyrexian Espionage",
		Completeness:  CompletenessFull,
		Purpose:       game.Purpose{Draws: 2},
		OptionalCosts: []game.AdditionalCost{Kicker("{1}{B}")},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return DrawCards{Player: ctx.Controller(), N: 2}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				if !ctx.WasKicked() {
					return nil
				}
				return eachOpponentDiscardsOne(ctx.Game, ctx.Item)
			}),
		),
	})
}
