package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inferno — Instant {5}{R}{R} (The Dark, #5638):
//
//	"Inferno deals 6 damage to each creature and each player."
//
// Jaya Ballard, Task Mage's last row as a spell: one damage instance
// to every creature on the battlefield and every player still in the
// game, the caster included. Damage, not destruction, so prevention
// and indestructible work as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "69e2df4e-c7f3-4c3a-be5b-1c4afb29cead",
		Name:         "Inferno",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 6}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.B23DamageEachCreatureAndEachPlayer(ctx, 6)
		},
	})
}
