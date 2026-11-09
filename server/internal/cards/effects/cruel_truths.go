package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cruel Truths — Instant {3}{B} (Tarkir: Dragonstorm, #4463):
//
//	"Surveil 2, then draw two cards. You lose 2 life."
//
// Risky Research's text, so its resolution: the surveil prompt carries
// the keyed tail (riskyResearchThen), the cards drawn come from the
// library the player arranged and the life loss follows the draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2e68e3f7-b168-4541-b8a0-791a8a6ebde8",
		Name:         "Cruel Truths",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return Surveil{Player: item.Controller, N: 2, ThenTail: ctx.Tail(riskyResearchThen)}.Apply(ctx)
		},
	})
}
