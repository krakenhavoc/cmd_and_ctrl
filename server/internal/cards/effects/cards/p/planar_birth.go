package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Planar Birth — Sorcery {1}{W} (Urza's Saga, S82 OLD batch 729, #5624):
//
//	"Return all basic land cards from all graveyards to the battlefield
//	 tapped under their owners' control."
//
// The set is every basic land card in every graveyard as the spell resolves.
// They all enter together (CR 603.6a), tapped, under their owners' control
// (the return's zero Controller).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7a9d13b3-3823-463a-acd9-8b7a9d5e121f",
		Name:         "Planar Birth",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			ids := effects.AllGraveyardsCardIDs(ctx, func(c game.Card) bool { return c.IsLand() && effects.IsBasicLand(c) })
			return effects.ReturnFromGraveyardTogether{Targets: ids, Tapped: true}.Apply(ctx)
		},
	})
}
