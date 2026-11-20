package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Untamed Wilds — Sorcery {2}{G} (Portal, #5630):
//
//	"Search your library for a basic land card, put that card onto the battlefield, then shuffle."
//
// Rampant Growth's search, with the land entering untapped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b3b4c21d-f8d7-455f-be46-d5eb909d54df",
		Name:         "Untamed Wilds",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Lands: 1, LandsUntapped: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: IsBasicLand,
				Dest:      game.ZoneBattlefield,
				Limit:     1,
				Shuffle:   true,
				Reason:    "Untamed Wilds — a basic land",
			}.Apply(ctx)
		},
	})
}
