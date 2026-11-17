package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ephemeron — Creature — Illusion {4}{U}{U}, 4/4 (Exodus, #5626):
//
//	"Flying
//	 Discard a card: Return this creature to its owner's hand."
//
// Flying is a printed keyword; the bounce is Thalakos Scout's row, a
// discard cost (CR 602.2b) and the shared
// returnThisPermanentToOwnersHand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e4f28144-71c8-4393-b383-047529fe8878",
		Name:            "Ephemeron",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "Discard a card: Return this creature to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    DiscardACard(),
			Effect:  returnThisPermanentToOwnersHand,
		}},
	})
}
