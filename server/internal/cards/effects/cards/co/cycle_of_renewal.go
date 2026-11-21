package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cycle of Renewal — Instant — Lesson {2}{G} (EDHREC rank 1978):
//
//	"Sacrifice a land. Search your library for up to two basic land
//	 cards, put them onto the battlefield tapped, then shuffle."
//
// Roiling Regrowth with a Lesson subtype — instant-speed ramp that
// nets one land and two landfall triggers. "Up to two" is the
// searcher's prompt with a minimum of zero.
//
// "Sacrifice a land" is part of the effect, not a cost: see
// Roiling Regrowth, whose body this shares.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "77eda626-5cef-496d-b291-cfa1ee6fc1fb",
		Name:         "Cycle of Renewal",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.SacrificeALandThenFetchTwoTapped(ctx, "Cycle of Renewal", effects.CycleOfRenewalFetch)
		},
	})
}
