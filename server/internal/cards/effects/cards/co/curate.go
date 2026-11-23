package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Curate — Instant {1}{U}:
//
//	"Surveil 2.
//	 Draw a card."
//
// Consider's shape with a deeper surveil: the draw rides the surveil's
// keyed tail, so it happens after the player has answered the prompt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "75e8907d-f726-4a63-ba8b-c0fd4936734e",
		Name:         "Curate",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			c := ctx.Controller()
			return effects.Surveil{Player: c, N: 2, ThenTail: game.KeyedTail{Ref: effects.ThenDrawACard, Player: c}}.Apply(ctx)
		},
	})
}
