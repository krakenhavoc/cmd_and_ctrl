package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fierce Witchstalker — Creature — Wolf {2}{G}{G}, 4/4:
//
//	"Trample
//	 When this creature enters, create a Food token."
//
// Trample rides PrintedKeywords; the Food is the shared FoodToken template.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ad69f48e-c387-4376-a04f-37d3992c7946",
		Name:            "Fierce Witchstalker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Fierce Witchstalker — create a Food token",
				Do(CreateToken{Template: FoodToken(), N: 1})),
		},
	})
}
