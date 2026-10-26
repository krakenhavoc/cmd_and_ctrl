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
// "Sacrifice a land" is part of the EFFECT, not a cost (the Entish
// Restoration pattern, #2876): the caster picks a land as the spell
// resolves and the search follows whether or not a land was
// sacrificed. A countered spell costs no land, and the spell can be
// cast with no land at all. Cycle of Renewal shares the body.
func init() {
	Register(Spec{
		OracleID:     "65d63518-3261-40b5-87b8-19c152b29ee3",
		Name:         "Roiling Regrowth",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return sacrificeALandThenFetchTwoTapped(ctx, "Roiling Regrowth", roilingRegrowthFetch)
		},
	})
}

// sacrificeALandThenFetchTwoTapped is "Sacrifice a land. Search your
// library for up to two basic land cards, put them onto the battlefield
// tapped, then shuffle." (Roiling Regrowth, Cycle of Renewal).
func sacrificeALandThenFetchTwoTapped(ctx *Context, name string, then game.PromptRunThen) error {
	return ctx.SacrificeThen(ctx.Controller(),
		sacrificeSpec("a land", Land()),
		name+" — sacrifice a land",
		1, then)
}

// roilingRegrowthFetch and cycleOfRenewalFetch are the search once the land
// has gone, keyed (#4290): one key per card, because the prompt names it.
var (
	roilingRegrowthFetch = fetchTwoBasicsTapped("prompt-run/roiling-regrowth-fetch", "Roiling Regrowth")
	cycleOfRenewalFetch  = fetchTwoBasicsTapped("prompt-run/cycle-of-renewal-fetch", "Cycle of Renewal")
)

// fetchTwoBasicsTapped registers "search your library for up to two basic
// land cards, put them onto the battlefield tapped, then shuffle" under key,
// for the card called name.
func fetchTwoBasicsTapped(key, name string) game.PromptRunThen {
	return RunThen(key, func(ctx *Context, _ game.RunDone) error {
		return SearchLibrary{
			Player:        ctx.Controller(),
			Predicate:     IsBasicLand,
			Dest:          game.ZoneBattlefield,
			Limit:         2,
			Shuffle:       true,
			TappedOnEntry: true,
			Reason:        name + " — up to two basic lands onto the battlefield tapped",
		}.Apply(ctx)
	})
}
