package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "d12e374d-ed09-4d5e-bb7a-6c7d9ea80a1e",
		Name:         "Brimstone Volley",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(DamageToTarget(0, 3)),
		Targets:      TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if b11CreaturesDiedThisTurn(ctx.Game) > 0 {
				return damageToFirstTarget(5)(item, ctx)
			}
			return damageToFirstTarget(3)(item, ctx)
		},
	})
}
