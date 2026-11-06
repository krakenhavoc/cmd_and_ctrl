package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skirmish Rhino — Creature — Rhino {W}{B}{G}, 3/4 (Tarkir: Dragonstorm,
// #4463):
//
//	"Trample
//	 When this creature enters, each opponent loses 2 life and you gain 2
//	 life."
//
// Sneering Shadewriter's drain: every opponent loses 2 life, then you
// gain 2, a fixed amount however many opponents there are.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "40eb7f43-7cee-41a1-a4d1-f34ee125ccd1",
		Name:            "Skirmish Rhino",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Skirmish Rhino — each opponent loses 2 life and you gain 2 life",
				Do(EachOpponentLosesLife{N: 2}, GainLife{Amount: 2})),
		},
	})
}
