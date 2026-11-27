package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flame Wave — Sorcery {3}{R}{R}{R}{R} (Stronghold, S82 OLD batch 747, #5627):
//
//	"Flame Wave deals 4 damage to target player or planeswalker and each
//	 creature that player or that planeswalker's controller controls."
//
// Bonfire of the Damned's body at a printed 4: the player is the target, or the
// controller of the targeted planeswalker, read before any damage lands; the
// creatures are snapshotted first and every point is one simultaneous damage
// instance (CR 608.2). With its one target illegal the spell does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cee8ec5f-e36d-4bca-9b19-cb33113835b1",
		Name:         "Flame Wave",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayerOrPlaneswalker(),
		Purpose: game.Purpose{
			Targets: game.ForTargets(effects.DamageToTarget(0, 4)),
			Sweep:   game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 4, OpponentsOnly: true, Partial: true},
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DamageToPlayerOrWalkerAndTheirCreatures(ctx, 4)
		},
	})
}
