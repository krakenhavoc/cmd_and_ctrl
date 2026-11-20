package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Silverback Shaman — Creature — Ape Shaman {3}{G}{G}, 5/4 (Core Set 2020, #5518):
//
//	"Trample
//	 When this creature dies, draw a card."
//
// Buzz Bots' dies trigger with a printed trample.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fc011947-b496-400e-99b8-b368068ba79b",
		Name:            "Silverback Shaman",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Silverback Shaman — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
