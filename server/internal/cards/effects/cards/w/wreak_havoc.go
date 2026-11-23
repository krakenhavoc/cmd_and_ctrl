package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "b1d2519e-7340-4ea4-a62e-cf0740d4e57a",
		Name:            "Wreak Havoc",
		Completeness:    effects.CompletenessFull,
		CantBeCountered: true,
		Targets:         effects.TargetPermanent("target artifact or land", effects.Or(effects.Artifact(), effects.Land())),
		Purpose:         game.Purpose{Answers: game.AnswerRemove},
		OnResolve:       effects.DestroyTheTargetPermanent,
	})
}
