package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stream of Acid — Sorcery {2}{B}{B} (Starter 1999, #5620):
//
//	"Destroy target land or nonblack creature."
//
// Befoul without "It can't be regenerated": one target clause, a land or
// a nonblack creature, judged at announce and again at resolution (CR
// 608.2b). A black land creature is still a land, so it can be chosen,
// and a regeneration shield saves the target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1b9c9083-62c3-46fd-a2a0-6de7de88a3c3",
		Name:         "Stream of Acid",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target land or nonblack creature", Or(Land(), And(Creature(), NonBlack()))),
		OnResolve:    destroyTheTargetPermanent,
	})
}
