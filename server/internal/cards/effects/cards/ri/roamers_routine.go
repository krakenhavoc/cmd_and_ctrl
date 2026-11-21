package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Roamer's Routine — Sorcery {2}{G} (#3931):
//
//	"Search your library for a basic land card, put it onto the
//	 battlefield tapped, then shuffle.
//	 Harmonize {4}{G}"
//
// Rampant Growth's search, and harmonize (CR 702.180).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "cf2355f3-11d2-4b03-b663-9fad6b07e4b2",
		Name:             "Roamer's Routine",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Harmonize("{4}{G}")},
		Purpose:          game.Purpose{Lands: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:        ctx.Controller(),
				Predicate:     effects.IsBasicLand,
				Dest:          game.ZoneBattlefield,
				Limit:         1,
				Shuffle:       true,
				TappedOnEntry: true,
				Reason:        "Roamer's Routine — a basic land card",
			}.Apply(ctx)
		},
	})
}
