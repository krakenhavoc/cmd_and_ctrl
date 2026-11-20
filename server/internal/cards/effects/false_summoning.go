package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// False Summoning — Instant {1}{U} (Portal Second Age, #5625):
//
//	"Counter target creature spell."
//
// Essence Scatter's text: the picker offers only creature spells, the
// clause is judged again at resolution (CR 608.2b), and the body is the
// shared counterTheTargetSpell. An artifact creature spell is a creature
// spell.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4f891c68-c959-4210-94e5-94a8e487d5ef",
		Name:         "False Summoning",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // a counterspell, as Negate declares
		Targets:      TargetSpell("target creature spell", Creature()),
		OnResolve:    counterTheTargetSpell,
	})
}
