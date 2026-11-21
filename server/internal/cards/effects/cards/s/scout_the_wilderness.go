package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "422aca0b-8e8f-4774-9a20-6f7a1cae967e",
		Name:         "Scout the Wilderness",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{Lands: 1,
			Kicked: &game.Purpose{Tokens: 2, CreatureTokens: 2}},
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{1}{W}")},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.RampantGrowthSearch()(ctx.Item, ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.WasKicked() {
					return nil
				}
				return effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier"), N: 2}.Apply(ctx)
			}),
		),
	})
}
