package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Searing Wind — Instant {8}{R} (Prophecy, #5617):
//
//	"Searing Wind deals 10 damage to any target."
//
// Lightning Bolt's target and the shared damageToFirstTarget body at
// 10. A target that became illegal in response means the spell doesn't
// resolve (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b196045d-ece1-46ac-a647-b65f04225c10",
		Name:         "Searing Wind",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 10)), Answers: game.AnswerRemove},
		OnResolve:    effects.DamageToFirstTargetStep(10),
	})
}
