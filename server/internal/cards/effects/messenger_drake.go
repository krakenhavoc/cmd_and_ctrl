package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Messenger Drake — Creature — Drake {3}{U}{U}, 3/3 (Magic 2014, #5563):
//
//	"Flying
//	 When this creature dies, draw a card."
//
// Kingfisher's dies trigger on a 3/3 flier: the player who last controlled it
// draws.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4c62dce6-dc1f-4d2d-95dd-295810ada6a6",
		Name:            "Messenger Drake",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Messenger Drake — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
