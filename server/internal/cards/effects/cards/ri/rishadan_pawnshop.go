package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rishadan Pawnshop — Artifact {2} (Mercadian Masques, S82 OLD b681, #5619):
//
//	"{2}, {T}: Shuffle target nontoken permanent you control into its owner's
//	 library."
//
// Unravel the Aether's tuck for a permanent you control that is not a token:
// it goes into its OWNER's library (a stolen permanent goes to its owner) and
// that library is shuffled. A target no longer legal does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9eef1d16-4405-40e0-8b8b-008e08aac4ac",
		Name:         "Rishadan Pawnshop",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Shuffle target nontoken permanent you control into its owner's library.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target nontoken permanent you control", effects.YouControl(), effects.Not(effects.IsTokenPredicate())),
			Effect:  effects.S82OldB377ShuffleTargetIntoOwnersLibrary,
		}},
	})
}
