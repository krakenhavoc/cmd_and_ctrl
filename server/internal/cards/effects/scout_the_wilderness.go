package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scout the Wilderness — {2}{G} Sorcery (Dominaria United):
//
//	"Kicker {1}{W} (You may pay an additional {1}{W} as you cast this
//	 spell.)
//	 Search your library for a basic land card, put it onto the
//	 battlefield tapped, then shuffle. If this spell was kicked, create
//	 two 1/1 white Soldier creature tokens."
//
// The search is Rampant Growth's. The tokens are made after it
// (ResolveSteps), and a library with no basic land still makes them.
func init() {
	Register(Spec{
		OracleID:     "422aca0b-8e8f-4774-9a20-6f7a1cae967e",
		Name:         "Scout the Wilderness",
		Completeness: CompletenessFull,
		Purpose: game.Purpose{Lands: 1,
			Kicked: &game.Purpose{Tokens: 2, CreatureTokens: 2}},
		OptionalCosts: []game.AdditionalCost{Kicker("{1}{W}")},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return rampantGrowthSearch()(ctx.Item, ctx)
			}),
			Instruction(func(ctx *Context) error {
				if !ctx.WasKicked() {
					return nil
				}
				return CreateToken{Template: TokenCard("1/1 white Soldier"), N: 2}.Apply(ctx)
			}),
		),
	})
}
