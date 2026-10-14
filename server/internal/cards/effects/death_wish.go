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
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (TakeFromOutside{
				Question: "Death Wish — put a card you own from outside the game into your hand, or none",
			}).Apply(ctx); err != nil {
				return err
			}
			// "You lose half your life, rounded up." Life loss, not
			// damage (CR 119.3), so no prevention sees it.
			if p := ctx.PlayerByID(item.Controller); p != nil {
				if n := game.HalfLifeRoundedUp(p.Life); n > 0 {
					if err := ctx.Game.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -n); err != nil {
						return err
					}
				}
			}
			// "Exile Death Wish." The spell moves itself (#489).
			return ctx.Game.ExileCardForEffect(item.SourceCardID)
		},
	})
}
