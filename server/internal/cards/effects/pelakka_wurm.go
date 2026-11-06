package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "d36075c2-de66-4202-9217-b1102a2bc14b",
		Name:            "Pelakka Wurm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Pelakka Wurm — gain 7 life", Do(GainLife{Amount: 7})),
			WhenThisDies("Pelakka Wurm — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
