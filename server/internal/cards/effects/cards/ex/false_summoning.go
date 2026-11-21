package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "4f891c68-c959-4210-94e5-94a8e487d5ef",
		Name:         "False Summoning",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // a counterspell, as Negate declares
		Targets:      effects.TargetSpell("target creature spell", effects.Creature()),
		OnResolve:    effects.CounterTheTargetSpell,
	})
}
