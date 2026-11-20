package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Neutralizing Blast — Instant {1}{U} (Fate Reforged,
// #5552):
//
//	"Counter target multicolored spell."
//
// Null Elemental Blast's counter bullet as a whole spell: "multicolored" is
// the spell's effective colour count, two or more, read as it is targeted
// and again as this resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e42ca67d-c9c3-499a-a7b7-63dc2dafdeab",
		Name:         "Neutralizing Blast",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetSpell("target multicolored spell", Multicolored()),
		OnResolve:    counterTheTargetSpell,
	})
}
