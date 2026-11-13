package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rack and Ruin — Instant {2}{R} (Urza's Legacy, #5623):
//
//	"Destroy two target artifacts."
//
// Two different artifacts (CR 115.3 forbids one object twice for one
// "target" word), destroyed together as one event by destroyEachLegalTarget.
// One that left or stopped being an artifact in response is skipped, and the
// other is still destroyed (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e04552f4-41c1-40f3-996e-de796c9967e4",
		Name:         "Rack and Ruin",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("two target artifacts", Artifact()).WithCount(2, 2),
		OnResolve:    destroyEachLegalTarget,
	})
}
