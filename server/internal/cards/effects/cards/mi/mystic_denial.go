package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mystic Denial — Instant {1}{U}{U} (Portal, #5630):
//
//	"Counter target creature or sorcery spell."
//
// Envelop's body, for a creature or sorcery spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d2bd23a6-4f77-4d6e-bf8f-339cb7a4184d",
		Name:         "Mystic Denial",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // a counterspell, as Negate declares
		Targets:      effects.TargetSpell("target creature or sorcery spell", effects.Or(effects.Creature(), effects.Sorcery())),
		OnResolve:    effects.CounterTheTargetSpell,
	})
}
