package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Artillerize — Instant {3}{R} (New Phyrexia, #5575):
//
//	"As an additional cost to cast this spell, sacrifice an artifact or creature.
//	 Artillerize deals 5 damage to any target."
//
// The sacrifice is an additional cost paid at cast (CR 601.2h,
// SacrificeCost), and Lightning Bolt's damage at 5.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "d14a8d90-13c0-4eb0-ab7b-86edb2c7fb96",
		Name:           "Artillerize",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("an artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
		Targets:        effects.TargetAny(),
		Purpose:        game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 5)), Answers: game.AnswerRemove},
		OnResolve:      effects.DamageToFirstTargetStep(5),
	})
}
