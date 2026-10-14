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
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (TakeFromOutside{
				Match:    func(c game.Card) bool { return c.IsCreature() || c.IsLand() },
				Question: "Living Wish — reveal a creature or land card you own from outside the game, or none",
				Reveal:   true,
			}).Apply(ctx); err != nil {
				return err
			}
			// "Exile Living Wish." The spell moves itself, so the resolution
			// frame does not put it into the graveyard afterwards (#489).
			return ctx.Game.ExileCardForEffect(item.SourceCardID)
		},
	})
}
