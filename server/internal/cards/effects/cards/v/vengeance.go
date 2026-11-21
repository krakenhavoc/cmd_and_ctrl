package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vengeance — Sorcery {3}{W} (Portal, #5630):
//
//	"Destroy target tapped creature."
//
// Swift Response's shape as a sorcery: the tapped-creature clause
// (tappedPermanent) is judged at announce and again at resolution
// (CR 608.2b), so a creature that untapped in response is not
// destroyed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1d001145-5d14-43a9-bf3b-3ce5c20b2a46",
		Name:         "Vengeance",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target tapped creature", effects.TappedPermanent()),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
