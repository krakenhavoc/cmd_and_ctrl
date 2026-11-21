package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Extinguish — Instant {1}{U} (Portal Second Age, #5625):
//
//	"Counter target sorcery spell."
//
// Envelop at {1}{U}: the clause is judged at announce and again at
// resolution (CR 608.2b), and the body is the shared
// counterTheTargetSpell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b2c9f074-57ca-4709-976d-f432f632483f",
		Name:         "Extinguish",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // a counterspell, as Negate declares
		Targets:      effects.TargetSpell("target sorcery spell", effects.Sorcery()),
		OnResolve:    effects.CounterTheTargetSpell,
	})
}
