package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghostly Visit — Sorcery {2}{B} (Portal Three Kingdoms, #5622):
//
//	"Destroy target nonblack creature."
//
// Hand of Death's shape (destroyTheTargetPermanent). Destroy moves it to
// its owner's graveyard (CR 701.8a); regeneration can still replace it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "73b6694a-47d4-417a-af9b-470da14b5b0e",
		Name:         "Ghostly Visit",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target nonblack creature", effects.NonBlack()),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
