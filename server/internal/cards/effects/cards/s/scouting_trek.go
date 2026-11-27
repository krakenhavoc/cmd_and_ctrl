package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scouting Trek — Sorcery {1}{G} (Invasion, S82 #5616):
//
//	"Search your library for any number of basic land cards, reveal those
//	 cards, then shuffle and put them on top."
//
// An any-number search, so the controller may take none; the cards found are
// revealed and put on top after the shuffle (they are basic lands, so their
// order is not a choice that matters).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "929c231a-12d9-4f1a-9849-64a8b5a65f20",
		Name:         "Scouting Trek",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:    item.Controller,
				Predicate: effects.IsBasicLand,
				Dest:      game.ZoneLibrary,
				ToTop:     true,
				Unbounded: true,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Scouting Trek — search for any number of basic land cards",
				Source:    item.SourceCardID,
			}.Apply(ctx)
		},
	})
}
