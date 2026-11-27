package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Harmonic Convergence — Instant {2}{G}:
//
//	"Put all enchantments on top of their owners' libraries."
//
// Every enchantment on the battlefield, whoever controls it, leaves at once for
// the top of its owner's library. An owner of two or more arranges their own
// in any order (CR 401.4); a token ceases to exist.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "400adeec-11b4-4a9c-9466-dcab524d87ea",
		Name:         "Harmonic Convergence",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepEnchantments, How: game.SweepBounce}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for _, c := range effects.MatchingBattlefield(ctx, effects.Enchantment()) {
				ids = append(ids, c.InstanceID)
			}
			return effects.OldB718TuckOnTopByOwner(ctx, ids, "Harmonic Convergence — put your enchantments on top of your library in any order")
		},
	})
}
