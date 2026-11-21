package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Expel — Instant {2}{W} (Strixhaven: School of Mages, #5502):
//
//	"Exile target tapped creature."
//
// Assassinate's "target tapped creature" clause with an exile where the
// destroy was. A creature that untaps in response is no longer a legal
// target and the spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "58776b25-794e-4a87-bc71-a77c025b9b03",
		Name:         "Expel",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target tapped creature", effects.TappedPermanent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ExileFirstLegalCardTarget(ctx.Game, item)
		},
	})
}
