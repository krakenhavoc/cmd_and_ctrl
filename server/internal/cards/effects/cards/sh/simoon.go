package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Simoon — Instant {R}{G} (Invasion, S82 #5616):
//
//	"Simoon deals 1 damage to each creature target opponent controls."
//
// The creatures are read as the spell resolves and dealt their damage as one
// instance; a target opponent who is no longer legal makes the spell do
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "128e201a-f520-4917-a1a4-2f3836f1f92d",
		Name:         "Simoon",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		Purpose: game.Purpose{
			Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, OpponentsOnly: true},
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.FirstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetPlayer {
				return nil
			}
			return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.ControlledBy(t.ID)), 1)
		},
	})
}
