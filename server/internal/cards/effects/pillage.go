package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pillage — Sorcery {1}{R}{R} (Alliances, #5634):
//
//	"Destroy target artifact or land. It can't be regenerated."
//
// A destruction (CR 701.8a) that no regeneration shield replaces. If the
// target has left by the time Pillage resolves, it does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0b137853-7cb9-424b-8285-12938991eafb",
		Name:         "Pillage",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target artifact or land", Or(Artifact(), Land())),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve:    destroyTheTargetPermanentNoRegen,
	})
}
