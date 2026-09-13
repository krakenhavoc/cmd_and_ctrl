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
// for.
//
// S22: the searcher picks both cards. The second search is CHAINED
// off the first via Then — it cannot run on the line below, because
// the first search now returns while its prompt is still open, and
// a second prompt opened at that moment would offer a card the
// player is in the middle of taking.
//
// "Up to two" degrades the way it reads: a library with one basic
// gives the battlefield half and finds nothing for the hand half; a
// library with none no-ops both, and either way the player may
// decline.
//
// The land enters the battlefield TAPPED because Cultivate says so
// (SearchLibrary.TappedOnEntry). Since #263 the fetched land's own
// enters-tapped clause runs as well, so a Cultivated checkland is
// tapped for the printed reason on top of this one.
func init() {
	Register(Spec{
		OracleID:     "8b755881-a72d-4e21-a369-d2924eb4585a",
		Name:         "Cultivate",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"If your library holds only one basic land, it is always put onto the battlefield tapped — you can't choose to put it into your hand instead."},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			controller := ctx.Controller()
			source := ctx.Source()
			return SearchLibrary{
				Player:        controller,
				Source:        source,
				Predicate:     IsBasicLand,
				Dest:          game.ZoneBattlefield,
				Limit:         1,
				Reveal:        true,
				Shuffle:       false,
				TappedOnEntry: true,
				Reason:        "Cultivate — basic land onto the battlefield tapped",
				Then: func(g *game.Game, _ []uuid.UUID) error {
					return g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
						Player:  controller,
						Source:  source,
						Pred:    IsBasicLand,
						Dest:    game.ZoneHand,
						Limit:   1,
						Reveal:  true,
						Shuffle: true,
						Reason:  "Cultivate — basic land into your hand",
					})
				},
			}.Apply(ctx)
		},
	})
}
