package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brimstone Volley — Instant {2}{R} (Innistrad, #5572):
//
//	"Brimstone Volley deals 3 damage to any target.
//	 Morbid — Brimstone Volley deals 5 damage instead if a creature
//	 died this turn."
//
// Morbid is read as the spell resolves, off the turn tally of creatures
// that died (Tragic Slip's b11CreaturesDiedThisTurn): 5 damage if any
// creature died this turn, 3 otherwise.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d12e374d-ed09-4d5e-bb7a-6c7d9ea80a1e",
		Name:         "Brimstone Volley",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		Targets:      effects.TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if effects.B11CreaturesDiedThisTurn(ctx.Game) > 0 {
				return effects.DamageToFirstTargetStep(5)(item, ctx)
			}
			return effects.DamageToFirstTargetStep(3)(item, ctx)
		},
	})
}
