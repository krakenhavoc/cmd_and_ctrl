package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cultivate — "Search your library for up to two basic land cards,
// reveal those cards, put one onto the battlefield tapped and the
// other into your hand, then shuffle."
//
// Two sequential searches rather than one, because the two halves go
// to different zones and the primitive has one destination. The
// first defers its shuffle so the second still sees the library;
// the second shuffles, which is the single shuffle the rules ask
// for. The second is CHAINED off the first via Then — it cannot run
// on the line below, because the first search returns while its
// prompt is still open, and a second prompt opened at that moment
// would offer a card the player is in the middle of taking.
//
// "Up to two" (#2942): BOTH halves are Optional, which forces the
// prompt even when only one match is left, so the searcher can take
// two, one (to the battlefield or to the hand, their choice of half)
// or none. Before, the second half took the last basic without
// asking, and a declined first half still forced a pick on the
// second. With exactly one basic in the library this subsumes the old
// lone-basic confirm (#2883): the first prompt offers it for the
// battlefield and, if declined, the second offers it for the hand. An
// empty library no-ops both halves but still shuffles.
//
// The land enters the battlefield TAPPED because Cultivate says so
// (SearchLibrary.TappedOnEntry). Since #263 the fetched land's own
// enters-tapped clause runs as well, so a Cultivated checkland is
// tapped for the printed reason on top of this one.
func init() {
	Register(Spec{
		OracleID:     "8b755881-a72d-4e21-a369-d2924eb4585a",
		Name:         "Cultivate",
		Purpose:      game.Purpose{Lands: 1, Tutors: 1},
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return cultivateSearch(ctx, "Cultivate")
		},
	})
}

// cultivateSearch is the shared "up to two basic land cards, one onto
// the battlefield tapped and the other into your hand" resolution,
// used by Cultivate and Kodama's Reach (#2942).
func cultivateSearch(ctx *Context, name string) error {
	controller := ctx.Controller()
	source := ctx.Source()
	return SearchLibrary{
		Player:        controller,
		Source:        source,
		Predicate:     IsBasicLand,
		Dest:          game.ZoneBattlefield,
		Limit:         1,
		Optional:      true,
		Reveal:        true,
		Shuffle:       false,
		TappedOnEntry: true,
		Reason:        name + " — basic land onto the battlefield tapped",
		Then: func(g *game.Game, _ []uuid.UUID) error {
			return g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
				Player:   controller,
				Source:   source,
				Pred:     IsBasicLand,
				Dest:     game.ZoneHand,
				Limit:    1,
				Optional: true,
				Reveal:   true,
				Shuffle:  true,
				Reason:   name + " — basic land into your hand",
			})
		},
	}.Apply(ctx)
}
