package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Swift Response — Instant {1}{W}:
//
//	"Destroy target tapped creature."
//
// "Tapped" is part of the target clause, so it is checked as the spell
// is cast and again as it resolves (CR 608.2b): a creature untapped in
// response is no longer a legal target and survives.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5820d760-8ed7-4abf-a959-c06ed86a3730",
		Name:         "Swift Response",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // #4734: destroys a permanent
		Targets:      effects.TargetCreature("target tapped creature", effects.TappedPermanent()),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
