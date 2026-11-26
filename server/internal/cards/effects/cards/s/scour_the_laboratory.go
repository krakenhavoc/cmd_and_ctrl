package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scour the Laboratory — Instant {4}{U}{U} (Eldritch Moon, S82 OLD, #5541):
//
//	"Delirium — This spell costs {2} less to cast if there are four or more
//	 card types among cards in your graveyard.
//	 Draw three cards."
//
// Rowdy Research's draw three behind Drag to the Roots' delirium discount
// (CR 601.2f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "983c5c59-806c-4430-adec-c91a099c8d41",
		Name:         "Scour the Laboratory",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 3},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(2, "Delirium — This spell costs {2} less to cast if there are four or more card types among cards in your graveyard.", effects.YouHaveDelirium()),
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: item.Controller, N: 3}.Apply(ctx)
		},
	})
}
