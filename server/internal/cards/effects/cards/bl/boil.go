package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boil — Instant {3}{R} (Eighth Edition, #5607):
//
//	"Destroy all Islands."
//
// Every land with the Island subtype, basic or not, whoever controls it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3a51485b-31b2-4fa5-824a-a919f9d28ca8",
		Name:         "Boil",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.Subtype("Island")}.Apply(ctx)
		},
	})
}
