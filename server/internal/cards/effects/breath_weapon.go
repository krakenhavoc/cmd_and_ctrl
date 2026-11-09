package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Breath Weapon — Instant {2}{R}:
//
//	"Breath Weapon deals 2 damage to each non-Dragon creature."
//
// Damage, not destruction: indestructible and prevention work as
// printed, and the deaths are the state-based sweep's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f3dd1f6d-f7d7-4358-8139-7495404f29c7",
		Name:         "Breath Weapon",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return damageEachMatching(ctx, And(Creature(), Not(cmdrB5Dragon())), 2)
		},
	})
}
