package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ghostly Visit — Sorcery {2}{B} (Portal Three Kingdoms, #5622):
//
//	"Destroy target nonblack creature."
//
// Hand of Death's shape (destroyTheTargetPermanent). Destroy moves it to
// its owner's graveyard (CR 701.8a); regeneration can still replace it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "73b6694a-47d4-417a-af9b-470da14b5b0e",
		Name:         "Ghostly Visit",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target nonblack creature", NonBlack()),
		OnResolve:    destroyTheTargetPermanent,
	})
}
