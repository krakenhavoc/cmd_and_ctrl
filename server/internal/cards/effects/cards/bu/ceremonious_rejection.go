package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ceremonious Rejection — Instant {U} (Kaladesh, #5539):
//
//	"Counter target colorless spell."
//
// Artifact Blast's counter with a colorless target spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6aa5aa3d-fb46-47e5-bcf1-51a28e0f9974",
		Name:         "Ceremonious Rejection",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // a counterspell
		Targets:      effects.TargetSpell("target colorless spell", effects.Colorless()),
		OnResolve:    effects.CounterTheTargetSpell,
	})
}
