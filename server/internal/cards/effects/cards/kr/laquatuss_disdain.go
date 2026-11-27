package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Laquatus's Disdain — Instant {1}{U} (Judgment, #5611):
//
//	"Counter target spell cast from a graveyard.
//	 Draw a card."
//
// The target is a spell whose cast announcement named a graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "68f330ca-8279-46c5-900d-a9b443676104",
		Name:         "Laquatus's Disdain",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target spell cast from a graveyard", effects.OldB621CastFromAGraveyard()),
		OnResolve:    effects.ResolveSteps(effects.OldB621CounterTheTargetSpell, effects.DrawCards{N: 1}),
	})
}
