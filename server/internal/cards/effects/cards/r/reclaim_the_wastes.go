package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reclaim the Wastes — Sorcery {G} (Zendikar Rising, S82 OLD #5508):
//
//	"Kicker {3} (You may pay an additional {3} as you cast this spell.)
//	 Search your library for a basic land card, reveal it, put it into your
//	 hand, then shuffle. If this spell was kicked, search your library for
//	 two basic land cards instead of one."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "ce6f3c67-8806-416d-9156-6f5bb63d5f4c",
		Name:          "Reclaim the Wastes",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{3}")},
		Purpose:       game.Purpose{Tutors: 1, Lands: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			n := 1
			if ctx.WasKicked() {
				n = 2
			}
			return effects.SearchLibrary{
				Player:    item.Controller,
				Predicate: effects.IsBasicLand,
				Dest:      game.ZoneHand,
				Limit:     n,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Reclaim the Wastes — basic land cards",
			}.Apply(ctx)
		},
	})
}
