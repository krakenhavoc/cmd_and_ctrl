package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boundless Realms — Sorcery {6}{G} (Magic 2013):
//
//	"Search your library for up to X basic land cards, where X is the number of
//	 lands you control, put them onto the battlefield tapped, then shuffle."
//
// Harvest Season's search with X counted as the spell resolves over the lands
// you control. X of zero is still a search and still a shuffle.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e6acafb7-856f-4125-a908-a620635b8ba3",
		Name:         "Boundless Realms",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := effects.B02CountLandsControlledBy(ctx.Game, ctx.Controller())
			match := effects.IsBasicLand
			if x <= 0 {
				match = func(game.Card) bool { return false }
			}
			return effects.SearchLibrary{
				Player:        ctx.Controller(),
				Predicate:     match,
				Dest:          game.ZoneBattlefield,
				Limit:         x,
				UpTo:          true,
				Shuffle:       true,
				TappedOnEntry: true,
				Reason:        "Boundless Realms — up to X basic lands, tapped",
			}.Apply(ctx)
		},
	})
}
