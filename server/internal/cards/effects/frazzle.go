package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Frazzle — Instant {3}{U} (Guildpact, #5599):
//
//	"Counter target nonblue spell."
//
// Remove Soul's counter (counterTheTargetSpell) with a colour predicate:
// a spell that is blue, a multicoloured one included, is not a legal
// target. Colour is read off the spell as it is on the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "55c4df81-9513-424b-aa58-10190797bba7",
		Name:         "Frazzle",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetSpell("target nonblue spell", Not(OfColor("U"))),
		OnResolve:    counterTheTargetSpell,
	})
}
