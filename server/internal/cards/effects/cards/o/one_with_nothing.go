package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// One with Nothing — Instant {B} (Saviors of Kamigawa, #5601):
//
//	"Discard your hand."
//
// Every card in the caster's hand is discarded at once, as Peer Past the
// Veil's first step does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ba995901-bb32-45ed-8af8-f98784b5a9b0",
		Name:         "One with Nothing",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			_, err := effects.DiscardWholeHand(ctx.Game, ctx.Controller())
			return err
		},
	})
}
