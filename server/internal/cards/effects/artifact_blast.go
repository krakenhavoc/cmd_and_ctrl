package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Artifact Blast — Instant {R} (Antiquities, #5640):
//
//	"Counter target artifact spell."
//
// Annul's body with the narrower clause: the picker offers only
// artifact spells, the engine refuses anything else at announce, and the
// target is checked again as Artifact Blast resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2d4aedc5-31c5-4281-98e1-b0c2233c3c8a",
		Name:         "Artifact Blast",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // a counterspell
		Targets:      TargetSpell("target artifact spell", Artifact()),
		OnResolve:    counterTheTargetSpell,
	})
}
