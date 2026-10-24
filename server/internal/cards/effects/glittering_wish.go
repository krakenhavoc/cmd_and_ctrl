package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glittering Wish — Sorcery {G}{W}:
//
//	"You may reveal a multicolored card you own from outside the game
//	 and put it into your hand. Exile Glittering Wish."
//
// The wish with a colour clause (ADR 0191): TakeFromOutside matching
// the two-or-more-colours test, the same one Multicolored() applies to
// the board.
func init() {
	Register(Spec{
		OracleID:     "45c29ac0-f9ca-4ab9-b93e-8b678509bcd7",
		Name:         "Glittering Wish",
		Completeness: CompletenessFull,
		// The exile waits for the question the take asks (CR 608.2c, #4155).
		OnResolve: ResolveSteps(
			TakeFromOutside{
				Pick:     glitteringWishTake,
				Question: "Glittering Wish — reveal a multicolored card you own from outside the game, or none",
			},
			Instruction(func(ctx *Context) error {
				// "Exile Glittering Wish." The spell moves itself (#489).
				return ctx.Game.ExileCardForEffect(ctx.Item.SourceCardID)
			}),
		),
	})
}

// glitteringWishTake is the wish's keyed pick (#4256), so a table waiting on
// it is a restore point.
var glitteringWishTake = registerOutsideTake("choose-cards/glittering-wish-take",
	func(c game.Card) bool { return len(c.EffectiveColors()) >= 2 }, true)
