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
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (TakeFromOutside{
				Match:    func(c game.Card) bool { return len(c.EffectiveColors()) >= 2 },
				Question: "Glittering Wish — reveal a multicolored card you own from outside the game, or none",
				Reveal:   true,
			}).Apply(ctx); err != nil {
				return err
			}
			// "Exile Glittering Wish." The spell moves itself (#489).
			return ctx.Game.ExileCardForEffect(item.SourceCardID)
		},
	})
}
