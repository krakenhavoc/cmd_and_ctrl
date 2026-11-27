package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Poison Arrow — Sorcery {4}{B}{B} (Portal Three Kingdoms, S82 OLD b713, #5622):
//
//	"Destroy target nonblack creature. You gain 3 life."
//
// The shared destroy-then-gain-life body (Cursebreak, Sanctify): an illegal
// target means no life either (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6d920c99-8a85-4b20-97c2-7aa0c79ae70a",
		Name:         "Poison Arrow",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target nonblack creature", effects.NonBlack()),
		OnResolve:    effects.DestroyTargetThenGainLife(3),
	})
}
