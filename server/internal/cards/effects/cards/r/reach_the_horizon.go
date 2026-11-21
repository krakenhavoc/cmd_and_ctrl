package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reach the Horizon — Sorcery {3}{G} (Final Fantasy, #4462):
//
//	"Search your library for up to two basic land cards and/or Town cards
//	 with different names, put them onto the battlefield tapped, then
//	 shuffle."
//
// Kodama's Reach's search with Prishe's Wanderings' wider net (a basic
// land card or a Town card, prishesWanderingsFinds) and Tiamat's set rule
// over the picks: two cards with different names, so two Forests are not
// a legal pair but a Forest and a Plains are. A stated quality may find
// fewer (CR 701.23b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "14bc29ee-4f84-4e62-8ede-5741b99414c9",
		Name:         "Reach the Horizon",
		Purpose:      game.Purpose{Lands: 2},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:        ctx.Controller(),
				Predicate:     effects.PrishesWanderingsFinds,
				Dest:          game.ZoneBattlefield,
				Limit:         2,
				Shuffle:       true,
				TappedOnEntry: true,
				Rule:          effects.SearchRuleDifferentNames,
				Reason:        "Reach the Horizon — up to two basic land cards and/or Town cards with different names, tapped",
			}.Apply(ctx)
		},
	})
}
