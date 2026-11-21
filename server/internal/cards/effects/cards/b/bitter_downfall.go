package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bitter Downfall — Instant {3}{B} (The Hobbit Commander):
//
//	"This spell costs {3} less to cast if it targets a creature that was
//	 dealt damage this turn.
//	 Destroy target creature. Its controller loses 2 life."
//
// #4454, S76. The reduction reads the announced target
// (CostsLessIfItTargets) with the shared DealtDamageThisTurn predicate.
// The creature's controller is read before the destroy, so the life loss
// still lands on them when the creature is regenerated or dies, and the
// whole spell does nothing when the target has left (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "19c22e08-29b6-4b3d-bf6e-28a5add620bb",
		Name:         "Bitter Downfall",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessIfItTargets(3, "This spell costs {3} less to cast if it targets a creature that was dealt damage this turn.",
				effects.And(effects.Creature(), effects.DealtDamageThisTurn())),
		},
		OnResolve: effects.JmpB2DestroyTargetCreatureItsControllerLosesTwo,
	})
}
