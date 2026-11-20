package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arborback Stomper — Creature — Beast {3}{G}{G}, 5/4 (Kaladesh, #5539):
//
//	"Trample
//	 When this creature enters, you gain 5 life."
//
// Trample is printed; the life gain is Bramble Wurm's enters trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a3f35e0f-bf7a-42bf-b612-ac3af3a33727",
		Name:            "Arborback Stomper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Arborback Stomper — you gain 5 life", Do(GainLife{Amount: 5})),
		},
	})
}
