package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Open the Gates — Sorcery {G} (Ravnica Remastered):
//
//	"Search your library for a basic land card or Gate card, reveal it, put
//	 it into your hand, then shuffle."
//
// A Gate card is any card with the Gate subtype (a land type), read off the
// searched card's own type line.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f38e43b6-9a8c-4f6a-bd1c-d5c4913a28a1",
		Name:         "Open the Gates",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:    item.Controller,
				Source:    item.SourceCardID,
				Predicate: func(c game.Card) bool { return effects.IsBasicLand(c) || c.HasSubtype("Gate") },
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Open the Gates — a basic land card or Gate card",
			}.Apply(ctx)
		},
	})
}
