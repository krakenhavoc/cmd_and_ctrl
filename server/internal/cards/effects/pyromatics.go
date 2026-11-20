package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pyromatics — Instant {1}{R} (Guildpact, #5599):
//
//	"Replicate {1}{R} (When you cast this spell, copy it for each time
//	 you paid its replicate cost. You may choose new targets for the
//	 copies.)
//	 Pyromatics deals 1 damage to any target."
//
// Replicate is Shattering Spree's pair (replicateMana and Replicate(),
// CR 702.56a), twenty payments a cast at most. Each copy deals 1 damage
// to its own target (damageTheClauseTarget).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "9c01e2a6-61e7-455c-b09f-8c0de2266782",
		Name:          "Pyromatics",
		Completeness:  CompletenessFull,
		Targets:       TargetAny(),
		Purpose:       ForTargets(DamageToTarget(0, 1)),
		OptionalCosts: []game.AdditionalCost{replicateMana("{1}{R}", 20)},
		Triggered:     []game.TriggeredAbility{Replicate()},
		OnResolve:     ResolveSteps(damageTheClauseTarget(0, 1)),
	})
}
