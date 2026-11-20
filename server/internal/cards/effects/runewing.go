package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Runewing — Creature — Bird {3}{U}, 2/2 (Return to Ravnica, #5566):
//
//	"Flying
//	 When this creature dies, draw a card."
//
// Kingfisher's dies trigger: it goes on the stack once the Bird is in the
// graveyard, and the player who last controlled it draws.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cc300649-ae50-47e7-ad0f-ffbdffddb989",
		Name:            "Runewing",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Runewing — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
