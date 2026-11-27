package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wayward Soul — Creature — Spirit {2}{U}{U}, 3/2 (Exodus, #5626):
//
//	"Flying
//	 {U}: Put this creature on top of its owner's library."
//
// Flying is a printed keyword. The ability is Gandalf the Grey's tuck
// as a whole row: TuckToLibraryForEffect puts the creature on top of
// its OWNER's library whoever controls it, and a commander's owner is
// asked whether it goes to the command zone instead (CR 903.9b). A
// Soul that has already left, or come back as a new object, is not
// moved (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "319b4715-c3eb-49bf-b8d2-4d41e601ebe4",
		Name:            "Wayward Soul",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}: Put this creature on top of its owner's library.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{U}"),
			Effect:  effects.PutThisPermanentOnTopOfOwnersLibrary,
		}},
	})
}
