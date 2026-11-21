package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Excoriate — Sorcery {3}{W} (Born of the Gods,
// #5559):
//
//	"Exile target tapped creature."
//
// Swallow Whole's exile without its additional cost: the tapped clause
// is on the target (tappedPermanent), checked again as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a9aedd96-1aee-480d-87fe-3ed94bd64599",
		Name:         "Excoriate",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target tapped creature", effects.TappedPermanent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ExileFirstLegalCardTarget(ctx.Game, item)
		},
	})
}
