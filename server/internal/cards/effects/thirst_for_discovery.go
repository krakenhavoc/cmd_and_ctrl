package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thirst for Discovery — Instant {2}{U} (EDHREC rank 3748):
//
//	"Draw three cards. Then discard two cards unless you discard a
//	 basic land card."
//
// Thirst for Knowledge with basic lands where that card has
// artifacts: three cards now, and the discard is either one basic
// land card or any two. The draws land first, then the discard prompt
// opens over the whole hand, so the drawn cards are among the choices
// (discardTwoUnlessOne).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1e05e6ef-14af-451d-9d54-e75b1f8871ab",
		Name:         "Thirst for Discovery",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return ctx.Game.DrawNThenForEffect(item.Controller, 3, game.DrawThen{
				Ref: drawThenDiscardTwoUnlessBasicLand, Player: item.Controller, Source: item.SourceCardID, N: 2,
			})
		},
	})
}
