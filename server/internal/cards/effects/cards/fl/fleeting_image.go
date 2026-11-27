package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fleeting Image — Creature — Illusion {2}{U}, 2/1 (Eighth Edition, S82 OLD
// batch 612, #5607):
//
//	"Flying
//	 {1}{U}: Return this creature to its owner's hand."
//
// Flying is the engine's keyword. The bounce is Quicksilver Wall's
// (ReturnThisPermanentToOwnersHand) for its controller alone: it is not a cost,
// so the creature stays until the ability resolves and can be answered, and a
// creature that has left, or come back as a new object (CR 400.7), is not moved.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "623f3efb-5431-42d4-98aa-9caac40e4e83",
		Name:            "Fleeting Image",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}: Return this creature to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{1}{U}"),
			Effect:  effects.ReturnThisPermanentToOwnersHand,
		}},
	})
}
