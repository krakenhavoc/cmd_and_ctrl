package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Desert Sandstorm — Sorcery {2}{R} (Portal Three Kingdoms, S82 #5622):
//
//	"Desert Sandstorm deals 1 damage to each creature."
//
// Pyroclasm's sweep at 1: the creatures are snapshotted before any damage lands
// and take it together; players and planeswalkers take none. Damage,
// not destruction, so indestructible and prevention work as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "230bf3a4-68a8-4553-b8d7-ad340cbd7371",
		Name:         "Desert Sandstorm",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DamageEachMatching(ctx, effects.Creature(), 1)
		},
	})
}
