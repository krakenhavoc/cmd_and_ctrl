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
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return sacrificeALandThenFetchTwoTapped(item, ctx, "Roiling Regrowth")
		},
	})
}

// sacrificeALandThenFetchTwoTapped is "Sacrifice a land. Search your
// library for up to two basic land cards, put them onto the battlefield
// tapped, then shuffle." (Roiling Regrowth, Cycle of Renewal).
func sacrificeALandThenFetchTwoTapped(item *game.StackItem, ctx *Context, name string) error {
	controller := ctx.Controller()
	return ctx.Game.PlayerSacrificesThenForEffect(
		ctx.Source(), controller,
		sacrificeSpec("a land", Land()),
		name+" — sacrifice a land",
		1,
		func(g *game.Game, _ game.PromptedSacrifices) error {
			return SearchLibrary{
				Player:        controller,
				Predicate:     IsBasicLand,
				Dest:          game.ZoneBattlefield,
				Limit:         2,
				Shuffle:       true,
				TappedOnEntry: true,
				Reason:        name + " — up to two basic lands onto the battlefield tapped",
			}.Apply(NewContext(g, item))
		})
}
