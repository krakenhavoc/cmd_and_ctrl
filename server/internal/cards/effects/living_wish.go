package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Living Wish — Sorcery {1}{G}:
//
//	"You may reveal a creature or land card you own from outside the game
//	 and put it into your hand. Exile Living Wish."
//
// Burning Wish's sibling (ADR 0191): the same primitive with the other
// type clause. A card that is both (Dryad Arbor) matches either way.
func init() {
	Register(Spec{
		OracleID:     "b9d186f5-1c92-4539-a5b5-e7fbd4b726fe",
		Name:         "Living Wish",
		Completeness: CompletenessFull,
		// The exile waits for the question the take asks (CR 608.2c, #4155).
		OnResolve: ResolveSteps(
			TakeFromOutside{
				Pick:     livingWishTake,
				Question: "Living Wish — reveal a creature or land card you own from outside the game, or none",
			},
			Instruction(func(ctx *Context) error {
				// "Exile Living Wish." The spell moves itself, so the resolution
				// frame does not put it into the graveyard afterwards (#489).
				return ctx.Game.ExileCardForEffect(ctx.Item.SourceCardID)
			}),
		),
	})
}

// livingWishTake is the wish's keyed pick (#4256), so a table waiting on
// it is a restore point.
var livingWishTake = registerOutsideTake("choose-cards/living-wish-take",
	func(c game.Card) bool { return c.IsCreature() || c.IsLand() }, true)
