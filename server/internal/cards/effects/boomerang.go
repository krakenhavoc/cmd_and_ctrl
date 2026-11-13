package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Boomerang — Instant {U}{U} (Legends, #5639):
//
//	"Return target permanent to its owner's hand."
//
// The shared bounceTheTarget body: the permanent goes to its OWNER's
// hand, whoever controls it, and a target gone by resolution does
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dc4a4996-108a-4aac-850f-2d9f76403446",
		Name:         "Boomerang",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target permanent"),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve:    bounceTheTarget,
	})
}
