package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tenderize — {1}{G} Instant (Teenage Mutant Ninja Turtles, #4457):
//
//	"Target creature you control deals damage equal to its power to
//	 target creature an opponent controls."
//
// Quarrel's text: two target clauses read by slot (oneSidedBite). The
// biter's power is read as the spell resolves, and nothing is dealt when
// either creature has become an illegal target (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a03fa584-ceb5-410c-9b80-f75d00215537",
		Name:         "Tenderize",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature an opponent controls", OpponentControls()),
		),
		OnResolve: oneSidedBite,
	})
}
