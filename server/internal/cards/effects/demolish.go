package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Demolish — Sorcery {3}{R} (Ravnica Remastered, #5948):
//
//	"Destroy target artifact or land."
//
// One target clause and the shared "Destroy target [permanent]." body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "08c6d2fd-d9ae-4f95-bb9d-7e16b1039814",
		Name:         "Demolish",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target artifact or land", Or(Artifact(), Land())),
		OnResolve:    destroyTheTargetPermanent,
	})
}
