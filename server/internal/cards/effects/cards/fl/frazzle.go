package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "55c4df81-9513-424b-aa58-10190797bba7",
		Name:         "Frazzle",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target nonblue spell", effects.Not(effects.OfColor("U"))),
		OnResolve:    effects.CounterTheTargetSpell,
	})
}
