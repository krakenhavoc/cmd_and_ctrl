package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Roiling Regrowth — Instant {2}{G} (EDHREC rank 1053):
//
//	"Sacrifice a land. Search your library for up to two basic land
//	 cards, put them onto the battlefield tapped, then shuffle."
//
// Harrow's tapped cousin, and the reason it is an instant: two
// landfall triggers at the end of an opponent's turn. "Up to two"
// is the searcher's prompt with a minimum of zero.
//
// Sandbox simplification, declared — the Entish Restoration posture:
// "Sacrifice a land" is modelled as an ADDITIONAL COST TO CAST, not a
// resolution-time action, because no sacrifice-then-continue prompt
// exists. Every observable difference runs the weaker way — the land
// is gone even if the spell is countered, and the spell cannot be
// cast with no land to sacrifice (printed, it can be cast to search
// for nothing).
func init() {
	Register(Spec{
		OracleID:       "65d63518-3261-40b5-87b8-19c152b29ee3",
		Name:           "Roiling Regrowth",
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
				Reason:        "Roiling Regrowth — up to two basic lands onto the battlefield tapped",
			}.Apply(ctx)
		},
	})
}
