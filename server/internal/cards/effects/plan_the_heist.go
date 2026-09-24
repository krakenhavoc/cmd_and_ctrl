package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Plan the Heist — {2}{U}{U} Sorcery:
//
//	"Surveil 3 if you have no cards in hand. Then draw three cards.
//	 Plot {3}{U} (You may pay {3}{U} and exile this card from your
//	 hand. Cast it as a sorcery on a later turn without paying its mana
//	 cost. Plot only as a sorcery.)"
//
// A plot proof card (#1342), and the one whose text rewards the
// keyword: plot it early, empty your hand, and cast it free later to
// surveil first. The hand is read on resolution — the spell itself is
// on the stack by then, so it never counts against its own condition.
//
// The draw rides Surveil's Then, never the next statement: surveil
// only queues the prompt, and a draw written after it would take the
// cards the player is still deciding about.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "a9471b33-b5b0-408a-9900-6964e5fe42da",
		Name:           "Plan the Heist",
		Completeness:   CompletenessFull,
		SpecialActions: []game.SpecialAction{Plot("{3}{U}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			c := ctx.Controller()
			draw := func(g *game.Game) error { return g.DrawNForEffect(c, 3) }
			if p := ctx.Game.PlayerByIDForEffect(c); p != nil && (p.Hand == nil || len(p.Hand.Cards) == 0) {
				return Surveil{Player: c, N: 3, Then: draw}.Apply(ctx)
			}
			return draw(ctx.Game)
		},
	})
}
