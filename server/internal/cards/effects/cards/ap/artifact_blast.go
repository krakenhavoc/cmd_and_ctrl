package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Artifact Blast — Instant {R} (Antiquities, #5640):
//
//	"Counter target artifact spell."
//
// Annul's body with the narrower clause: the picker offers only
// artifact spells, the engine refuses anything else at announce, and the
// target is checked again as Artifact Blast resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2d4aedc5-31c5-4281-98e1-b0c2233c3c8a",
		Name:         "Artifact Blast",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // a counterspell
		Targets:      effects.TargetSpell("target artifact spell", effects.Artifact()),
		OnResolve:    effects.CounterTheTargetSpell,
	})
}
