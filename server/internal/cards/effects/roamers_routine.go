package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:         "cf2355f3-11d2-4b03-b663-9fad6b07e4b2",
		Name:             "Roamer's Routine",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Harmonize("{4}{G}")},
		Purpose:          game.Purpose{Lands: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:        ctx.Controller(),
				Predicate:     IsBasicLand,
				Dest:          game.ZoneBattlefield,
				Limit:         1,
				Shuffle:       true,
				TappedOnEntry: true,
				Reason:        "Roamer's Routine — a basic land card",
			}.Apply(ctx)
		},
	})
}
