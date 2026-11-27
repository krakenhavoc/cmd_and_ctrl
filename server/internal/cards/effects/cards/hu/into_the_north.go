package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Into the North — Sorcery {1}{G} (Coldsnap):
//
//	"Search your library for a snow land card, put it onto the battlefield
//	 tapped, then shuffle."
//
// A snow land card is a land card with the snow supertype (CR 205.4g), so a
// Snow-Covered basic or a nonbasic snow land such as Glacial Fortress's
// cousins qualifies. A stated quality lets the search find nothing
// (CR 701.23b), and the library is shuffled either way.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "30aefabc-5fd2-451b-aa93-8e5a972c7691",
		Name:         "Into the North",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Lands: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:        ctx.Controller(),
				Predicate:     func(c game.Card) bool { return c.IsLand() && c.HasSupertype("snow") },
				Dest:          game.ZoneBattlefield,
				Limit:         1,
				Shuffle:       true,
				TappedOnEntry: true,
				Reason:        "Into the North — a snow land",
			}.Apply(ctx)
		},
	})
}
