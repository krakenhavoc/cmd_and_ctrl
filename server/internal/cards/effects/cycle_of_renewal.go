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
// Sandbox simplification, declared — the Roiling Regrowth / Entish
// Restoration posture: "Sacrifice a land" is modelled as an
// ADDITIONAL COST TO CAST, not a resolution-time action, because no
// sacrifice-then-continue prompt exists. Every observable difference
// runs the weaker way — the land is gone even if the spell is
// countered, and the spell cannot be cast with no land to sacrifice
// (printed, it can be cast to search for nothing).
func init() {
	Register(Spec{
		OracleID:       "77eda626-5cef-496d-b291-cfa1ee6fc1fb",
		Name:           "Cycle of Renewal",
		Completeness:   CompletenessCaveats,
		Caveats:        []string{"The land sacrifice is paid when you cast it, so you lose the land even if the spell is countered, and you can't cast it with no land to sacrifice."},
		AdditionalCost: SacrificeCost("a land", Land()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:        item.Controller,
				Predicate:     IsBasicLand,
				Dest:          game.ZoneBattlefield,
				Limit:         2,
				Shuffle:       true,
				TappedOnEntry: true,
				Reason:        "Cycle of Renewal — up to two basic lands onto the battlefield tapped",
			}.Apply(ctx)
		},
	})
}
