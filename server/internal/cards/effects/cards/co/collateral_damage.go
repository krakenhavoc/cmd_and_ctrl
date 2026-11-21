package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Collateral Damage — Instant {R} (Fate Reforged,
// #5552):
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Collateral Damage deals 3 damage to any target."
//
// Reckless Abandon's shape for 3, as an instant: the sacrifice is an
// additional cost paid as the spell is cast (CR 118.8), so it is gone even
// if the spell is countered, and the spell can't be cast without a creature
// to sacrifice.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "b89dd47e-0f65-41f0-ad23-cefc61e68978",
		Name:           "Collateral Damage",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("a creature", effects.Creature()),
		Targets:        effects.TargetAny(),
		Purpose:        effects.WithRiders(effects.ForTargets(effects.DamageToTarget(0, 3)), game.Purpose{Answers: game.AnswerSacOutlet}),
		OnResolve:      effects.DamageToFirstTargetStep(3),
	})
}
