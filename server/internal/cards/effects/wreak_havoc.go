package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wreak Havoc — Sorcery {2}{R}{G} (Guildpact, #5599):
//
//	"This spell can't be countered.
//	 Destroy target artifact or land."
//
// Plunder's target and body (destroyTheTargetPermanent) with Supreme
// Verdict's CantBeCountered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b1d2519e-7340-4ea4-a62e-cf0740d4e57a",
		Name:            "Wreak Havoc",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		Targets:         TargetPermanent("target artifact or land", Or(Artifact(), Land())),
		Purpose:         game.Purpose{Answers: game.AnswerRemove},
		OnResolve:       destroyTheTargetPermanent,
	})
}
