package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Abundant Harvest — Sorcery {G} (Strixhaven Mystical Archive, #5503):
//
//	"Choose land or nonland. Reveal cards from the top of your library
//	 until you reveal a card of the chosen kind. Put that card into your
//	 hand and the rest on the bottom of your library in a random order."
//
// Pavel Chekov, First Posting's choice and run, on a sorcery: the caster
// picks land or nonland as it resolves, then Page, Loose Leaf's
// reveal-until takes the first card of that kind to the hand and puts
// every other revealed card on the bottom in a random order. A library
// with no card of the chosen kind is revealed whole and all of it goes
// under; an empty library reveals nothing, and nobody loses, because
// revealing is not drawing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "34841228-31ac-48ac-b6d6-486b5e1e07dd",
		Name:         "Abundant Harvest",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return PickOption{
				Question: "Abundant Harvest — choose land or nonland",
				Options:  []game.ChoiceOption{{Label: "Land"}, {Label: "Nonland"}},
				ThenKey:  abundantHarvestChosen,
			}.Apply(ctx)
		},
	})
}

// abundantHarvestChosen reveals until a card of the chosen kind, takes it
// and puts the rest on the bottom in a random order. Nobody chose (the
// caster left the game): nothing happens. An on-disk key: never renamed,
// never reused.
var abundantHarvestChosen = OptionPickThen("option-pick/abundant-harvest-kind", func(ctx *Context, r game.OptionPicked) error {
	if r.Option == nil {
		return nil
	}
	wantLand := r.Index == 0
	kind := "nonland"
	if wantLand {
		kind = "land"
	}
	return socB30RevealUntilToHandRestBottom("Abundant Harvest — "+kind,
		func(c game.Card) bool { return c.IsLand() == wantLand })(ctx.Game, ctx.Item)
})
