package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Golden Wish — Sorcery:
//
//	"You may reveal an artifact or enchantment card you own from outside
//	 the game and put it into your hand. Exile Golden Wish."
//
// The wish whose clause is a pair of types (ADR 0191): TakeFromOutside
// matching an artifact or an enchantment, as Living Wish matches a
// creature or a land.
func init() {
	Register(Spec{
		OracleID:     "e5f25cbd-39e4-4806-8dba-9a552356696d",
		Name:         "Golden Wish",
		Completeness: CompletenessFull,
		// The exile waits for the question the take asks (CR 608.2c, #4155).
		OnResolve: ResolveSteps(
			TakeFromOutside{
				Pick:     goldenWishTake,
				Question: "Golden Wish — reveal an artifact or enchantment card you own from outside the game, or none",
			},
			Instruction(func(ctx *Context) error {
				// "Exile Golden Wish." The spell moves itself (#489).
				return ctx.Game.ExileCardForEffect(ctx.Item.SourceCardID)
			}),
		),
	})
}

// goldenWishTake is the wish's keyed pick (#4256), so a table waiting
// on it is a restore point.
var goldenWishTake = registerOutsideTake("choose-cards/golden-wish-take",
	func(c game.Card) bool { return c.IsArtifact() || c.IsEnchantment() }, true)
