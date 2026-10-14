package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cunning Wish — Instant {2}{U}:
//
//	"You may reveal an instant card you own from outside the game and
//	 put it into your hand. Exile Cunning Wish."
//
// Burning Wish's sibling (ADR 0191) with an instant clause: the same
// TakeFromOutside, matching instants. The spell is an instant itself,
// so it resolves in the instant window and exiles itself either way.
func init() {
	Register(Spec{
		OracleID:     "bdf80a55-b524-4d17-9be8-a0005403a0e7",
		Name:         "Cunning Wish",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (TakeFromOutside{
				Match:    func(c game.Card) bool { return c.IsInstant() },
				Question: "Cunning Wish — reveal an instant card you own from outside the game, or none",
				Reveal:   true,
			}).Apply(ctx); err != nil {
				return err
			}
			// "Exile Cunning Wish." The spell moves itself, so the
			// resolution frame does not put it into the graveyard (#489).
			return ctx.Game.ExileCardForEffect(item.SourceCardID)
		},
	})
}
