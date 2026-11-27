package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Preemptive Strike — Instant {1}{U} (Portal Three Kingdoms, S82 OLD b714, #5622):
//
//	"Counter target creature spell."
//
// Essence Scatter's text: the clause is enforced at announce and at
// resolution (CR 601.2c, 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "250f8642-9754-48fd-8f09-70ed13d7a42c",
		Name:         "Preemptive Strike",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target creature spell", effects.Creature()),
		OnResolve:    effects.CounterTheTargetSpell,
	})
}
