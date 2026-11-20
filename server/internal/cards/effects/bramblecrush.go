package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bramblecrush — Sorcery {2}{G}{G} (Innistrad, #5572):
//
//	"Destroy target noncreature permanent."
//
// Chain of Acid's target (any noncreature permanent) destroyed by the
// shared destroyTheTargetPermanent; an indestructible target survives.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "367282d6-956b-470d-a131-d22beff610b0",
		Name:         "Bramblecrush",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target noncreature permanent", Noncreature()),
		OnResolve:    destroyTheTargetPermanent,
	})
}
