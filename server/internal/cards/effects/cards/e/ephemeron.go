package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "e4f28144-71c8-4393-b383-047529fe8878",
		Name:            "Ephemeron",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Discard a card: Return this creature to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.DiscardACard(),
			Effect:  effects.ReturnThisPermanentToOwnersHand,
		}},
	})
}
