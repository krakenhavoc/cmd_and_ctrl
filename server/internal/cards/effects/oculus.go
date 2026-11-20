package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Oculus — Creature — Phyrexian Homunculus {1}{U}, 1/1 (Mirrodin Besieged,
// #5576):
//
//	"When this creature dies, you may draw a card."
//
// Solemn Simulacrum's dies half: an optional draw on the stack, so it can
// be declined.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e793ede9-044a-44e9-bc9c-99b0034503eb",
		Name:         "Oculus",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OptionalDraw(WhenThisDies("Oculus — draw a card", Do(DrawCards{N: 1})), "Oculus — draw a card?", 1),
		},
	})
}
