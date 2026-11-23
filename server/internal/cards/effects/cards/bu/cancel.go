package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cancel — Instant {1}{U}{U} (EDHREC rank 2058):
//
//	"Counter target spell."
//
// The three-mana Counterspell, and the reason every set has one.
// Counterspell's shape with a different cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7d00fb28-ea6c-49a9-b4af-ffb38860a9a7",
		Name:         "Cancel",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return effects.CounterTarget{StackID: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
