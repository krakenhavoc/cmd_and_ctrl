package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Time Reversal — Sorcery {3}{U}{U} (Magic 2011, #5579):
//
//	"Each player shuffles their hand and graveyard into their library,
//	 then draws seven cards. Exile Time Reversal."
//
// Echo of Eons' wheel (echoOfEonsResolve: every player's hand and
// graveyard go into their own library, and only once every shuffle has
// landed does each player draw seven), then Divergent Equation's
// self-exile. Time Reversal is on the stack while the graveyards are
// shuffled, so it is not among them, and it is exiled rather than put
// into its owner's graveyard (CR 608.2n). A copy has no card to exile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9a63f6fc-dcf8-4529-a598-ca3d511f29e2",
		Name:         "Time Reversal",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := echoOfEonsResolve(item, ctx); err != nil {
				return err
			}
			if item.IsCopy || sourceIsNewObject(ctx.Game, item) {
				return nil
			}
			return ctx.Game.ExileCardForEffect(item.SourceCardID)
		},
	})
}
