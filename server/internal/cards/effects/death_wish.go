package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Death Wish — Sorcery:
//
//	"You may put a card you own from outside the game into your hand.
//	 You lose half your life, rounded up. Exile Death Wish."
//
// Mastermind's Acquisition's outside mode with no reveal and a "may"
// (ADR 0191), plus a life payment. The life is lost whether or not a
// card is taken: the second sentence is not part of the "may". Half is
// read as the spell resolves; putting a card into the hand changes no
// life total, so the order of the clauses is not observable.
func init() {
	Register(Spec{
		OracleID:     "ff0ca1a4-da67-407c-a5c0-998a35eb7d8d",
		Name:         "Death Wish",
		Completeness: CompletenessFull,
		// The life loss and the exile wait for the question the take asks, and the exile for
		// a prompt the loss asks (CR 608.2c, #4155).
		OnResolve: ResolveSteps(
			TakeFromOutside{
				Question: "Death Wish — put a card you own from outside the game into your hand, or none",
			},
			Instruction(func(ctx *Context) error {
				// "You lose half your life, rounded up." Life loss, not
				// damage (CR 119.3), so no prevention sees it.
				if p := ctx.PlayerByID(ctx.Item.Controller); p != nil {
					if n := game.HalfLifeRoundedUp(p.Life); n > 0 {
						return ctx.Game.ChangePlayerLifeForEffect(ctx.Item.SourceCardID, ctx.Item.Controller, -n)
					}
				}
				return nil
			}),
			Instruction(func(ctx *Context) error {
				// "Exile Death Wish." The spell moves itself (#489).
				return ctx.Game.ExileCardForEffect(ctx.Item.SourceCardID)
			}),
		),
	})
}
