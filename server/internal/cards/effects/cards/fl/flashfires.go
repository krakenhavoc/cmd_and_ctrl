package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flashfires — Sorcery {3}{R} (Eighth Edition, #5607):
//
//	"Destroy all Plains."
//
// Every land with the Plains subtype, basic or not, whoever controls it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c281f436-8c77-48f7-b31c-d40cd7f9ed6a",
		Name:         "Flashfires",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.Subtype("Plains")}.Apply(ctx)
		},
	})
}
