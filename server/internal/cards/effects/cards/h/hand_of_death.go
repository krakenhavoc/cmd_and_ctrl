package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hand of Death — Sorcery {2}{B} (Portal, #5630):
//
//	"Destroy target nonblack creature."
//
// Doom Blade's shape as a sorcery: the shared destroy body
// (destroyTheTargetPermanent). Destroy moves it to its owner's
// graveyard (CR 701.8a); regeneration can still replace it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dc45b2e3-272b-479b-8e3b-36eead606a3a",
		Name:         "Hand of Death",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target nonblack creature", effects.NonBlack()),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
