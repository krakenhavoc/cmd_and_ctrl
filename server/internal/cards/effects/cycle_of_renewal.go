package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cycle of Renewal — Instant — Lesson {2}{G} (EDHREC rank 1978):
//
//	"Sacrifice a land. Search your library for up to two basic land
//	 cards, put them onto the battlefield tapped, then shuffle."
//
// Roiling Regrowth with a Lesson subtype — instant-speed ramp that
// nets one land and two landfall triggers. "Up to two" is the
// searcher's prompt with a minimum of zero.
//
// "Sacrifice a land" is part of the effect, not a cost: see
// Roiling Regrowth, whose body this shares.
func init() {
	Register(Spec{
		OracleID:     "77eda626-5cef-496d-b291-cfa1ee6fc1fb",
		Name:         "Cycle of Renewal",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return sacrificeALandThenFetchTwoTapped(ctx, "Cycle of Renewal", cycleOfRenewalFetch)
		},
	})
}
