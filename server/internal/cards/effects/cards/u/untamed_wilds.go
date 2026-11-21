package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Untamed Wilds — Sorcery {2}{G} (Portal, #5630):
//
//	"Search your library for a basic land card, put that card onto the battlefield, then shuffle."
//
// Rampant Growth's search, with the land entering untapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b3b4c21d-f8d7-455f-be46-d5eb909d54df",
		Name:         "Untamed Wilds",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Lands: 1, LandsUntapped: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: effects.IsBasicLand,
				Dest:      game.ZoneBattlefield,
				Limit:     1,
				Shuffle:   true,
				Reason:    "Untamed Wilds — a basic land",
			}.Apply(ctx)
		},
	})
}
