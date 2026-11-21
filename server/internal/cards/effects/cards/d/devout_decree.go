package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Devout Decree — Sorcery {1}{W} (Foundations, #4465):
//
//	"Exile target creature or planeswalker that's black or red. Scry 1."
//
// Refute Destiny's target with the colours swapped, resolved as Ray of
// Ruin does: the exile, then the scry, which waits for any prompt the
// exile raised (CR 608.2c). The colours are words of the card's text, so
// a text-changing effect reaches them (CR 612.2). A spell whose only
// target is illegal as it resolves doesn't resolve, scry and all (CR
// 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "92f8c8b1-1fd5-4e80-b8a9-51671937aaf2",
		Name:         "Devout Decree",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Scry: 1},
		Targets: effects.TargetPermanent("target creature or planeswalker that's black or red",
			effects.And(effects.Or(effects.Creature(), effects.Planeswalker()), effects.Or(effects.OfColor("B"), effects.OfColor("R")))),
		OnResolve: effects.ResolveSteps(effects.ExileChosenTarget{}, effects.Scry{N: 1}),
	})
}
