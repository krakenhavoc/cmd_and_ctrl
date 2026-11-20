package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deliberate — Instant {1}{U} (Zendikar Rising, #5508):
//
//	"Scry 2, then draw a card."
//
// Preordain's text at instant speed. The draw rides Scry.ThenTail, a
// registered tail, so it happens after the player has put the cards back
// (CR 701.22a) and the scry prompt is a restore point.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "362fc643-ab76-4160-a73e-73397464db40",
		Name:         "Deliberate",
		Purpose:      game.Purpose{Draws: 1, Scry: 2},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return Scry{Player: ctx.Controller(), N: 2, ThenTail: ctx.Tail(thenDrawACard)}.Apply(ctx)
		},
	})
}
