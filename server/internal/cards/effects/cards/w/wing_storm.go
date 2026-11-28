package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wing Storm — Sorcery {2}{G} (Prophecy, S82 OLD b670, #5617):
//
//	"Wing Storm deals damage to each player equal to twice the number of
//	 creatures that player controls with flying."
//
// Price of Progress's DamageEachPlayerByCount: every count is taken before any
// damage lands, through the layers, and the caster is a player too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c593f66d-0e9a-4d46-a9c6-ac7a70bf17ea",
		Name:         "Wing Storm",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DamageEachPlayerByCount(ctx, func(player uuid.UUID) int {
				return 2 * effects.CountControlled(ctx.Game, player, func(c game.Card) bool {
					return c.IsCreature() && game.HasKeyword(&c, "flying")
				})
			})
		},
	})
}
