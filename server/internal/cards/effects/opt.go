package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Opt — Instant {U}:
//
//	"Scry 1.
//	 Draw a card."
//
// Preordain at instant speed and one card shallower. The two lines
// are printed as separate sentences rather than joined by "then",
// but the order is the order they are printed in (CR 608.2c), so the
// draw still happens after the scry and the card left on top is the
// card drawn — which is the whole reason anyone plays it.
//
// The draw therefore goes in Scry.Then, not on the line after. Scry
// only QUEUES a prompt: a draw written as the next statement would
// resolve first, take one of the cards the player is still deciding
// about, and leave the prompt unanswerable because that card is no
// longer in the library to put back. Preordain's file records the
// same trap.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "713332c1-5bd8-400f-bfff-c1ca0697a043",
		Name:         "Opt",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			controller := ctx.Controller()
			return Scry{
				Player: controller,
				N:      1,
				Then: func(g *game.Game) error {
					return g.DrawNForEffect(controller, 1)
				},
			}.Apply(ctx)
		},
	})
}
