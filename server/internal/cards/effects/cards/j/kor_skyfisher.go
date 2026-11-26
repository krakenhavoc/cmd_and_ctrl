package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kor Skyfisher — Creature — Kor Soldier {1}{W}, 2/3 (Duel Decks: Elspeth vs.
// Tezzeret, #5578):
//
//	"Flying
//	 When this creature enters, return a permanent you control to its owner's
//	 hand."
//
// Flying is a printed keyword. The enters trigger is mandatory: its
// controller picks any permanent they control (the Skyfisher itself
// included, which is the usual reply when it has nothing else to return) as
// the trigger resolves (CR 608.2c), ReturnOneYouControl's shared body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "054a0193-861e-4644-ab0a-8aa33407abc8",
		Name:            "Kor Skyfisher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Kor Skyfisher — return a permanent you control to its owner's hand", effects.Do(effects.ReturnOneYouControl{
				Question: "Kor Skyfisher — return a permanent you control to its owner's hand",
			})),
		},
	})
}
