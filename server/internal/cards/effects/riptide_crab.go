package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Riptide Crab — Creature — Crab {1}{W}{U}, 1/3 (Invasion, #5616):
//
//	"Vigilance
//	 When this creature dies, draw a card."
//
// Vigilance is a printed keyword; the dies trigger is Buzz Bots' draw
// for its controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8ae8a583-9843-474a-bd46-8090df883621",
		Name:            "Riptide Crab",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Riptide Crab — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
