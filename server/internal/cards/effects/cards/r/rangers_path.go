package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ranger's Path — Sorcery {3}{G} (Magic 2013, #5568):
//
//	"Search your library for up to two Forest cards, put them onto the
//	 battlefield tapped, then shuffle."
//
// Skyshroud Claim with the lands entering tapped and no reveal. "A
// Forest card" is any land with the Forest subtype, and the searcher
// may find fewer (CR 701.23b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "259a023a-de55-4bab-9e24-01aebcf716eb",
		Name:         "Ranger's Path",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Lands: 2},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:        ctx.Controller(),
				Predicate:     effects.IsLandWithSubtype("forest"),
				Dest:          game.ZoneBattlefield,
				Limit:         2,
				TappedOnEntry: true,
				Shuffle:       true,
				Reason:        "Ranger's Path — up to two Forest cards",
			}.Apply(ctx)
		},
	})
}
