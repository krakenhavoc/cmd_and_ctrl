package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Choking Fumes — Instant {2}{W} (Mirrodin Besieged, S82 OLD batch 464,
// #5576):
//
//	"Put a -1/-1 counter on each attacking creature."
//
// Midnight Banshee's mass -1/-1 counter (Cmdr2B25CountersOnEach), over
// every creature that is attacking as the spell resolves, any player's.
// The set is read before the first counter goes on.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "51bd7bc4-12cd-434e-8691-3123a018e634",
		Name:         "Choking Fumes",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.Cmdr2B25CountersOnEach(false, effects.IsAttackingCreature())(ctx.Game, item)
		},
	})
}
