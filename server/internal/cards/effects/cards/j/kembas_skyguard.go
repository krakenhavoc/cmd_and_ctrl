package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kemba's Skyguard — Creature — Cat Knight {1}{W}{W}, 2/2 (Duel Decks:
// Elspeth vs. Tezzeret, #5578):
//
//	"Flying
//	 When this creature enters, you gain 2 life."
//
// Mandroid Squadron's enters lifegain on a flier.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4c8167af-2eb0-4f0a-a1b9-b61b65bac261",
		Name:            "Kemba's Skyguard",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Kemba's Skyguard — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}
