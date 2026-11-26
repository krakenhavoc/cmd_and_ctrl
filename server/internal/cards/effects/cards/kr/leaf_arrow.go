package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leaf Arrow — Instant {G} (Rise of the Eldrazi, S82 OLD batch 483, #5581):
//
//	"Leaf Arrow deals 3 damage to target creature with flying."
//
// The flying check reads the layered view, at announce and again as the
// spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d3d67bf0-2e9e-463f-8df8-990ab906fdec",
		Name:         "Leaf Arrow",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
		Purpose: game.Purpose{
			Answers: game.AnswerRemove,
			Targets: game.ForTargets(effects.DamageToTarget(0, 3)),
		},
		OnResolve: effects.DamageToFirstTargetStep(3),
	})
}
