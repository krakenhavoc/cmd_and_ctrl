package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Palace Familiar — Creature — Bird {1}{U}, 1/1 (Dragons of Tarkir,
// #5551):
//
//	"Flying
//	 When this creature dies, draw a card."
//
// Buzz Bots' dies trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "dab9e40f-b73a-4c55-8bb3-63b32c23f07b",
		Name:            "Palace Familiar",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Palace Familiar — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
