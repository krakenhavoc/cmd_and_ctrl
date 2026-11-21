package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Three Visits — Sorcery {1}{G}:
//
//	"Search your library for a Forest card, put it onto the
//	battlefield, then shuffle."
//
// Functionally identical to Nature's Lore, and played alongside it
// for exactly that reason. Untapped, so TappedOnEntry stays false.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1b882a0e-0ede-4d1a-bd1a-9b7cffbcde8e",
		Name:         "Three Visits",
		Purpose:      game.Purpose{Lands: 1, LandsUntapped: 1},
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: effects.IsLandWithSubtype("forest"),
				Dest:      game.ZoneBattlefield,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Three Visits — a Forest card",
			}.Apply(ctx)
		},
	})
}
