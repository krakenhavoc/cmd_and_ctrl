package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Purple-Crystal Crab — Creature — Crab {1}{U}, 1/1 (Foundations
// Jumpstart):
//
//	"When this creature dies, draw a card."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "df4569d7-ee75-4ba0-bbc3-7766ac686523",
		Name:         "Purple-Crystal Crab",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Purple-Crystal Crab — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
