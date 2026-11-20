package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mystic Denial — Instant {1}{U}{U} (Portal, #5630):
//
//	"Counter target creature or sorcery spell."
//
// Envelop's body, for a creature or sorcery spell.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d2bd23a6-4f77-4d6e-bf8f-339cb7a4184d",
		Name:         "Mystic Denial",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // a counterspell, as Negate declares
		Targets:      TargetSpell("target creature or sorcery spell", Or(Creature(), Sorcery())),
		OnResolve:    counterTheTargetSpell,
	})
}
