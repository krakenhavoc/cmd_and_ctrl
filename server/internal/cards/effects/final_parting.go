package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Final Parting — Sorcery {3}{B}{B} (EDHREC rank 2609):
//
//	"Search your library for two cards. Put one into your hand and
//	 the other into your graveyard. Then shuffle."
//
// A Demonic Tutor and an Entomb in one card. "Two cards, one to hand
// and one to graveyard" is asked as two one-card searches in that
// order — the first through the S22 chooser to hand, the second,
// opened from the first's continuation, to the graveyard — so the
// searcher looks at the whole library both times and puts each card
// exactly where the card says; the shuffle comes once, after the
// second. Neither search reveals: the card says neither "reveal"
// nor which card went where, and only the graveyard half is public
// by being there. A library with a single card puts it into the
// hand and finds nothing for the graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a5852994-d816-4e62-8a03-254223714544",
		Name:         "Final Parting",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			controller, source := ctx.Controller(), ctx.Source()
			return SearchLibrary{
				Player:  controller,
				Dest:    game.ZoneHand,
				Limit:   1,
				Shuffle: false,
				Reason:  "Final Parting — a card to put into your hand",
				Then: func(g *game.Game, _ []uuid.UUID) error {
					return g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
						Player:  controller,
						Source:  source,
						Dest:    game.ZoneGraveyard,
						Limit:   1,
						Shuffle: true,
						Reason:  "Final Parting — a card to put into your graveyard",
					})
				},
			}.Apply(ctx)
		},
	})
}
