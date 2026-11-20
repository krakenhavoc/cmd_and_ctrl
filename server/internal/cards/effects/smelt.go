package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Smelt — Instant {R} (Magic 2013, #5568):
//
//	"Destroy target artifact."
//
// Shatter at one mana: one target clause and the shared "Destroy target
// [permanent]." body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b5f4b732-099c-46e7-a03d-d8e4c78245b8",
		Name:         "Smelt",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target artifact", Artifact()),
		OnResolve:    destroyTheTargetPermanent,
	})
}
