package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stampeding Wildebeests — Creature — Antelope Beast {2}{G}{G}, 5/4
// (Visions, S82 OLD batch 786, #5631):
//
//	"Trample
//	 At the beginning of your upkeep, return a green creature you control to
//	 its owner's hand."
//
// Roaring Primadox's trigger with a green creature as the noun: mandatory,
// the controller picks as it resolves (not a target), and the Wildebeests
// is itself green and so a legal pick. With no green creature it does
// nothing. The colour is a word of the card's own text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c270c770-91ed-4540-b281-c4b60b943b34",
		Name:            "Stampeding Wildebeests",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Stampeding Wildebeests — return a green creature you control", effects.Do(effects.ReturnOneYouControl{
				Match:    func(c game.Card) bool { return c.IsCreature() && c.HasColorWord("G") },
				Question: "Stampeding Wildebeests — return a green creature you control to its owner's hand",
			})),
		},
	})
}
