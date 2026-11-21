package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aven of Enduring Hope — Creature — Bird Cleric {4}{W}, 3/3 (Hour of
// Devastation, #5534):
//
//	"Flying
//	 When this creature enters, you gain 3 life."
//
// Angel of Mercy's text: flying is a printed keyword and the life gain is
// an enters trigger on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e8291185-9d87-4aa2-b7ba-d4946e5a082f",
		Name:            "Aven of Enduring Hope",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Aven of Enduring Hope — you gain 3 life", effects.Do(effects.GainLife{Amount: 3})),
		},
	})
}
