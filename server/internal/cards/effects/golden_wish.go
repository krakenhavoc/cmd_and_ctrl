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
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (TakeFromOutside{
				Match:    func(c game.Card) bool { return c.IsArtifact() || c.IsEnchantment() },
				Question: "Golden Wish — reveal an artifact or enchantment card you own from outside the game, or none",
				Reveal:   true,
			}).Apply(ctx); err != nil {
				return err
			}
			// "Exile Golden Wish." The spell moves itself (#489).
			return ctx.Game.ExileCardForEffect(item.SourceCardID)
		},
	})
}
