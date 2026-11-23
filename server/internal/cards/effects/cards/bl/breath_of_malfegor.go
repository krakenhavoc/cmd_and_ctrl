package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Breath of Malfegor — Instant {3}{B}{R} (Alara Reborn, #5586):
//
//	"Breath of Malfegor deals 5 damage to each opponent."
//
// Boltwave's body at 5: one damage instance to every opponent (CR 615.8).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "81e6ade9-a38f-4655-b09c-9ca41a11face",
		Name:         "Breath of Malfegor",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DamageToEachOpponent(ctx.Game, item, 5)
		},
	})
}
