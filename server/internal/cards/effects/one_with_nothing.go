package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// One with Nothing — Instant {B} (Saviors of Kamigawa, #5601):
//
//	"Discard your hand."
//
// Every card in the caster's hand is discarded at once, as Peer Past the
// Veil's first step does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ba995901-bb32-45ed-8af8-f98784b5a9b0",
		Name:         "One with Nothing",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			_, err := discardWholeHand(ctx.Game, ctx.Controller())
			return err
		},
	})
}
