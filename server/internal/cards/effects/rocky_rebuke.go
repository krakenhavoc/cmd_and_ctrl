package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rocky Rebuke — Instant {1}{G}:
//
//	"Target creature you control deals damage equal to its power to
//	 target creature an opponent controls."
//
// Tenderize's one-sided fight: two target clauses read by slot, the
// biter's power read as the spell resolves, and nothing dealt if either
// creature has become an illegal target (CR 608.2b).
//
// No simplification. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:     "fd85db70-8616-42df-9dac-bad7923f88d7",
		Name:         "Rocky Rebuke",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature an opponent controls", OpponentControls()),
		),
		OnResolve: oneSidedBite,
	})
}
