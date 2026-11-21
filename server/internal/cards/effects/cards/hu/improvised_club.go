package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Improvised Club — Instant {1}{R} (The Hobbit Commander):
//
//	"As an additional cost to cast this spell, sacrifice an artifact or
//	 creature.
//	 Improvised Club deals 4 damage to any target."
//
// #4454, S76. Deadly Dispute's additional cost (paid with the spell
// already on the stack, CR 601.2h) and Lightning Bolt's damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "20c705e3-4141-41f8-ad73-bf953f0e1ddd",
		Name:           "Improvised Club",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("an artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
		Targets:        effects.TargetAny(),
		Purpose:        effects.ForTargets(effects.DamageToTarget(0, 4)),
		OnResolve:      effects.DamageToFirstTargetStep(4),
	})
}
