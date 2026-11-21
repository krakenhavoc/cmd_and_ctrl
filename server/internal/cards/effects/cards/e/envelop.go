package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Envelop — Instant {U} (Judgment, #5611):
//
//	"Counter target sorcery spell."
//
// Negate narrowed to sorceries: the clause is judged at announce and
// again at resolution (CR 601.2c, 608.2b), and the body is the shared
// counterTheTargetSpell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "30062dd0-c049-4872-8011-8b4810a3fa26",
		Name:         "Envelop",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // a counterspell, as Negate declares
		Targets:      effects.TargetSpell("target sorcery spell", effects.Sorcery()),
		OnResolve:    effects.CounterTheTargetSpell,
	})
}
