package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:      "efce3575-66ee-412f-af5b-961fd242bbde",
		Name:          "Reap and Sow",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Entwine("{1}{G}")},
		Modes: effects.Entwining(effects.ChooseOne(
			effects.ModeDoing("Destroy target land.", effects.TargetPermanent("target land", effects.Land()), effects.DestroyTheModesTarget),
			effects.ModeDoing("Search your library for a land card, put that card onto the battlefield, then shuffle.", nil,
				func(item *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.SearchLibrary{
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
