package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Acidic Soil — Sorcery {2}{R} (Urza's Saga):
//
//	"Acidic Soil deals damage to each player equal to the number of lands
//	 they control."
//
// Price of Progress counting every land: each seated player, the caster
// included, takes damage equal to the lands they control, all counted
// before any damage is dealt and dealt as one instance.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4ccc1ac5-db90-4fc9-9795-7afaf7466c3e",
		Name:         "Acidic Soil",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DamageEachPlayerByCount(ctx, func(player uuid.UUID) int {
				return effects.B02CountLandsControlledBy(ctx.Game, player)
			})
		},
	})
}
