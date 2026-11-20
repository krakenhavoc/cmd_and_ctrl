package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shared Roots — Sorcery — Lesson {1}{G} (EDHREC rank 1094):
//
//	"Search your library for a basic land card, put it onto the
//	 battlefield tapped, then shuffle."
//
// Rampant Growth with a Lesson subtype. The subtype matters only to
// "learn" (CR 701.48a, learn.go), which looks for it outside the game;
// the spell itself is the two-mana basic fetch, and the searcher picks
// the basic.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9e0fd3bf-f47a-4f06-8ff1-73f6bf5d1e03",
		Name:         "Shared Roots",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:        item.Controller,
				Predicate:     effects.IsBasicLand,
				Dest:          game.ZoneBattlefield,
				Limit:         1,
				Shuffle:       true,
				TappedOnEntry: true,
				Reason:        "Shared Roots — a basic land onto the battlefield tapped",
			}.Apply(ctx)
		},
	})
}
