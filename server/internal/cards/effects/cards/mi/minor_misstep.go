package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Minor Misstep — Instant {U} (Phyrexia: All Will Be One, #5479):
//
//	"Counter target spell with mana value 1 or less."
//
// Disdainful Stroke's shape: the mana value is a stack target predicate read
// on the stack (CR 202.3e), checked at announce and again at resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d91b3c6f-f8a1-4859-90ca-22cd79517552",
		Name:         "Minor Misstep",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target spell with mana value 1 or less", effects.ManaValueLE(1)),
		OnResolve:    effects.CounterTheTargetSpell,
	})
}
