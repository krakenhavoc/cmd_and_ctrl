package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chain Reaction — Sorcery {2}{R}{R}:
//
//	"Chain Reaction deals X damage to each creature, where X is the
//	 number of creatures on the battlefield."
//
// X is counted once, as the spell resolves, before any damage is
// dealt, and every creature takes that same X (a creature that dies
// to it does so at the state-based sweep afterwards, so the count
// does not shrink mid-resolution). It counts every creature on the
// battlefield, the caster's included. Damage, not destruction:
// indestructible and prevention shields work as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID: "086b2564-9114-4ba2-94fd-b490f98f38a7",
		Name:     "Chain Reaction",
		// ADR 0126 §6: the damage is the number of creatures, counted as it resolves, so it is declared as the destruction it usually is, an upper bound.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DamageEachMatching(ctx, effects.Creature(), len(effects.MatchingBattlefield(ctx, effects.Creature())))
		},
	})
}
