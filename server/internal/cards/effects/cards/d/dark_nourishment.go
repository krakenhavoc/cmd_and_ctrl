package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dark Nourishment — Instant {4}{B} (Ixalan, #5532):
//
//	"Dark Nourishment deals 3 damage to any target. You gain 3 life."
//
// Morbid Hunger's body without flashback: the damage, then the life
// gain, in printed order (CR 608.2c). The life is gained even if the
// target became illegal, as long as the spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "95650a60-6f1d-4594-8b53-7fd406212207",
		Name:         "Dark Nourishment",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 3)), LifeGain: 3, Answers: game.AnswerRemove},
		OnResolve:    effects.ResolveSteps(effects.DamageTheClauseTarget(0, 3), effects.GainLife{Amount: 3}),
	})
}
