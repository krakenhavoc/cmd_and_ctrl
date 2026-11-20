package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Oxidize — Instant {G} (Darksteel, #5605):
//
//	"Destroy target artifact. It can't be regenerated."
//
// Terror's no-regeneration destroy on an artifact target (CR 701.19c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0ff78353-a26e-4b5a-948d-1c3b41d2dd1f",
		Name:         "Oxidize",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target artifact", Artifact()),
		OnResolve:    destroyTheTargetPermanentNoRegen,
	})
}
