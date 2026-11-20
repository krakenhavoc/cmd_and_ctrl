package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ceremonious Rejection — Instant {U} (Kaladesh, #5539):
//
//	"Counter target colorless spell."
//
// Artifact Blast's counter with a colorless target spell.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6aa5aa3d-fb46-47e5-bcf1-51a28e0f9974",
		Name:         "Ceremonious Rejection",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // a counterspell
		Targets:      TargetSpell("target colorless spell", Colorless()),
		OnResolve:    counterTheTargetSpell,
	})
}
