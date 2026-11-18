package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "2249001c-08d5-4c0a-86f8-d97519a39f37",
		Name:         "Stoke the Flames",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		Targets:      TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 4)), Answers: game.AnswerRemove},
		OnResolve:    damageToFirstTarget(4),
	})
}
