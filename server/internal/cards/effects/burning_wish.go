package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Burning Wish — Sorcery {1}{R}:
//
//	"You may reveal a sorcery card you own from outside the game and put
//	 it into your hand. Exile Burning Wish."
//
// The first of the wishes on the sideboard seam (ADR 0191): a deck's
// sideboard is the player's pool of cards outside the game. TakeFromOutside
// asks for the sorcery, reveals it to the table and puts it into the hand;
// with no sorcery in the sideboard nothing is asked. Burning Wish exiles
// itself either way, as the last instruction.
func init() {
	Register(Spec{
		OracleID:     "13a884b5-8970-4d71-9082-a8b6c8db9a01",
		Name:         "Burning Wish",
		Completeness: CompletenessFull,
		// The exile waits for the question the take asks (CR 608.2c, #4155).
		OnResolve: ResolveSteps(
			TakeFromOutside{
				Pick:     burningWishTake,
				Question: "Burning Wish — reveal a sorcery card you own from outside the game, or none",
			},
			Instruction(func(ctx *Context) error {
				// "Exile Burning Wish." The spell moves itself, so the resolution
				// frame does not put it into the graveyard afterwards (#489).
				return ctx.Game.ExileCardForEffect(ctx.Item.SourceCardID)
			}),
		),
	})
}

// burningWishTake is the wish's keyed pick (#4256), so a table waiting on
// it is a restore point.
var burningWishTake = registerOutsideTake("choose-cards/burning-wish-take",
	func(c game.Card) bool { return c.IsSorcery() }, true)
