package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shard Volley — Instant {R} (Morningtide, #5592):
//
//	"As an additional cost to cast this spell, sacrifice a land.
//	 Shard Volley deals 3 damage to any target."
//
// Crop Rotation's sacrifice-a-land additional cost (CR 601.2h, paid as it
// is cast) on Lightning Strike's damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "3d19ab38-fb32-4f2b-b68c-28db4a6864c6",
		Name:           "Shard Volley",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("a land", effects.Land()),
		Targets:        effects.TargetAny(),
		Purpose:        game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 3)), Answers: game.AnswerRemove},
		OnResolve:      effects.DamageToFirstTargetStep(3),
	})
}
