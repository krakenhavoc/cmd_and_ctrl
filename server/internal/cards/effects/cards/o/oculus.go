package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "e793ede9-044a-44e9-bc9c-99b0034503eb",
		Name:         "Oculus",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OptionalDraw(effects.WhenThisDies("Oculus — draw a card", effects.Do(effects.DrawCards{N: 1})), "Oculus — draw a card?", 1),
		},
	})
}
