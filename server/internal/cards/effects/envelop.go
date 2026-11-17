package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "30062dd0-c049-4872-8011-8b4810a3fa26",
		Name:         "Envelop",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // a counterspell, as Negate declares
		Targets:      TargetSpell("target sorcery spell", Sorcery()),
		OnResolve:    counterTheTargetSpell,
	})
}
