package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crush — Instant {R} (Mirrodin Besieged,
// #5576):
//
//	"Destroy target noncreature artifact."
//
// Shatter's destroy with the clause narrowed to a noncreature artifact,
// judged at announce and again at resolution (CR 608.2b): an artifact that
// became a creature in response is no longer a legal target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ad4dda1e-2236-45bf-b2f3-f63e790d837d",
		Name:         "Crush",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // #4734: destroys a permanent
		Targets:      TargetPermanent("target noncreature artifact", Artifact(), Noncreature()),
		OnResolve:    destroyTheTargetPermanent,
	})
}
