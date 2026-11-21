package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Natural Connection — Instant {2}{G} (Battle for Zendikar, #5547):
//
//	"Search your library for a basic land card, put it onto the
//	 battlefield tapped, then shuffle."
//
// Shared Roots' search at instant speed: only a basic land card is
// offered, it enters tapped, and the library is shuffled afterwards.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dc85c659-f2ed-4b13-807c-d853f74de410",
		Name:         "Natural Connection",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:        item.Controller,
				Predicate:     effects.IsBasicLand,
				Dest:          game.ZoneBattlefield,
				Limit:         1,
				Shuffle:       true,
				TappedOnEntry: true,
				Reason:        "Natural Connection — a basic land onto the battlefield tapped",
			}.Apply(ctx)
		},
	})
}
