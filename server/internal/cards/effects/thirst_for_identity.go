package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thirst for Identity — Instant {2}{U} (Lorwyn Eclipsed, #4458):
//
//	"Draw three cards. Then discard two cards unless you discard a
//	 creature card."
//
// Thirst for Discovery with creature cards: three cards now, then the
// discard is either one creature card or any two. The draws land first,
// so the drawn cards are among the choices (discardTwoUnlessOne).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "045ece1e-23aa-4122-8726-86b2af4345ba",
		Name:         "Thirst for Identity",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return ctx.Game.DrawNThenForEffect(item.Controller, 3, game.DrawThen{
				Ref: stdB63ThirstForIdentityDiscard, Player: item.Controller, Source: item.SourceCardID, N: 2,
			})
		},
	})
}
