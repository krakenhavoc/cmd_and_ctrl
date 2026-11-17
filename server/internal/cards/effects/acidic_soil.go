package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "4ccc1ac5-db90-4fc9-9795-7afaf7466c3e",
		Name:         "Acidic Soil",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return damageEachPlayerByCount(ctx, func(player uuid.UUID) int {
				return b02CountLandsControlledBy(ctx.Game, player)
			})
		},
	})
}
