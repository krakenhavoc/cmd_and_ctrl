package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thwart — Instant {2}{U}{U} (Mercadian Masques, #5837):
//
//	"You may return three Islands you control to their owner's hand rather
//	 than pay this spell's mana cost.
//	 Counter target spell."
//
// Gush's alternative cost with three Islands (ReturnInstead over a clause
// of exactly three, #5837), returned to their owners' hands as a cost with
// Thwart already on the stack (CR 118.9, 601.2h, 400.3), so they are gone
// even if Thwart is countered. "Islands" is the land type (CR 205.3i). The
// body is Counterspell's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "615a149a-2869-4b18-bbbe-9ec1825049d7",
		Name:         "Thwart",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetSpell("target spell"),
		AlternativeCosts: []game.AlternativeCost{
			ReturnInstead(
				"Return three Islands you control to their owner's hand",
				PermanentYouControl("three Islands you control", HasSubtype("Island")).WithCount(3, 3),
				"three Islands you control",
			),
		},
		OnResolve: counterTheTargetSpell,
	})
}
