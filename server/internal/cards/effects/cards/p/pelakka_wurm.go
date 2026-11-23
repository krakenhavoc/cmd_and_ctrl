package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pelakka Wurm — Creature — Wurm {4}{G}{G}{G}, 7/7 (Foundations, #4465):
//
//	"Trample
//	 When this creature enters, you gain 7 life.
//	 When this creature dies, draw a card."
//
// Filigree Familiar's two triggers with 7 life; trample rides
// PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d36075c2-de66-4202-9217-b1102a2bc14b",
		Name:            "Pelakka Wurm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Pelakka Wurm — gain 7 life", effects.Do(effects.GainLife{Amount: 7})),
			effects.WhenThisDies("Pelakka Wurm — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
