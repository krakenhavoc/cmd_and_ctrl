package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stoke the Flames — Instant {2}{R}{R}:
//
//	"Convoke
//	 Stoke the Flames deals 4 damage to any target."
//
// Flame Lash's any-target clause with convoke (CR 702.51a) as the tap
// cost. A target gone by resolution takes nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2249001c-08d5-4c0a-86f8-d97519a39f37",
		Name:         "Stoke the Flames",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Convoke(),
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 4)), Answers: game.AnswerRemove},
		OnResolve:    effects.DamageToFirstTargetStep(4),
	})
}
