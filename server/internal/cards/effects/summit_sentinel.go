package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Summit Sentinel — Creature — Elemental Soldier {1}{U}, 1/3 (Lorwyn
// Eclipsed, #4458):
//
//	"When this creature dies, draw a card."
//
// A leaves-the-battlefield trigger that fires when the Sentinel dies
// (CR 700.4), written with the footprinted draw primitive.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e7b88da3-7dd1-45f8-b555-98ad838b2917",
		Name:         "Summit Sentinel",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Summit Sentinel — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
