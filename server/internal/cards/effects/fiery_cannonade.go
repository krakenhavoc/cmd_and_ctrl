package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fiery Cannonade — Instant {2}{R} (Ixalan, #5532):
//
//	"Fiery Cannonade deals 2 damage to each non-Pirate creature."
//
// Breath Weapon's sweep with Pirate for Dragon: every creature that
// isn't a Pirate as it resolves is dealt 2 damage at once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7e108b31-2ea3-4129-b324-a2d31c78bbe6",
		Name:         "Fiery Cannonade",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return damageEachMatching(ctx, And(Creature(), Not(HasSubtype("Pirate"))), 2)
		},
	})
}
