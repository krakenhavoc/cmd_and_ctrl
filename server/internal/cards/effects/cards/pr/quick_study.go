package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quick Study — Instant {2}{U} (EDHREC rank 2644):
//
//	"Draw two cards."
//
// Divination at instant speed — the cantrip-and-a-half every blue
// deck holds up on an opponent's end step. One primitive.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7bff8d4a-1c7d-48b8-b3e3-737dd6f01823",
		Name:         "Quick Study",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: item.Controller, N: 2}.Apply(ctx)
		},
	})
}
