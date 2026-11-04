package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// rampantGrowthSearch is Rampant Growth's whole resolution: "Search your
// library for a basic land card, put it onto the battlefield tapped, then
// shuffle." Rampant Growth and Studious First-Year's prepared spell (the
// same printed card) both call it.
func rampantGrowthSearch() func(*game.StackItem, *Context) error {
	return func(item *game.StackItem, ctx *Context) error {
		return SearchLibrary{
			Player:        ctx.Controller(),
			Predicate:     IsBasicLand,
			Dest:          game.ZoneBattlefield,
			Limit:         1,
			Shuffle:       true,
			TappedOnEntry: true,
			Reason:        "Rampant Growth — a basic land",
		}.Apply(ctx)
	}
}
