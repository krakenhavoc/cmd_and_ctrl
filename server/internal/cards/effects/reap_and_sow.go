package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reap and Sow — Sorcery {3}{G}:
//
//	"Choose one —
//	 • Destroy target land.
//	 • Search your library for a land card, put that card onto the
//	   battlefield, then shuffle.
//	 Entwine {1}{G} (Choose both if you pay the entwine cost.)"
//
// The land fetched enters untapped (nothing says tapped). A search for
// "a land card" names a quality, so it may find nothing (CR 701.23b).
// Entwined (CR 702.42, #3866) the destroy comes first, then the search
// (CR 702.42b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "efce3575-66ee-412f-af5b-961fd242bbde",
		Name:          "Reap and Sow",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Entwine("{1}{G}")},
		Modes: Entwining(ChooseOne(
			ModeDoing("Destroy target land.", TargetPermanent("target land", Land()), DestroyTheModesTarget),
			ModeDoing("Search your library for a land card, put that card onto the battlefield, then shuffle.", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return c.IsLand() },
						Dest:      game.ZoneBattlefield,
						Limit:     1,
						Shuffle:   true,
						Reason:    "Reap and Sow — a land card",
					}.Apply(ctx)
				}),
		)),
	})
}
