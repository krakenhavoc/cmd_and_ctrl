package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "dc45b2e3-272b-479b-8e3b-36eead606a3a",
		Name:         "Hand of Death",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target nonblack creature", NonBlack()),
		OnResolve:    destroyTheTargetPermanent,
	})
}
