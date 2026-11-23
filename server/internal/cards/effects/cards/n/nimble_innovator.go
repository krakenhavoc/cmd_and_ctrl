package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nimble Innovator — Creature — Vedalken Artificer {3}{U}, 2/2 (Kaladesh, #5539):
//
//	"When this creature enters, draw a card."
//
// Cloudkin Seer's enters draw, on the stack so it can be answered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "30929713-47dc-40f0-883c-4bd41aa399d5",
		Name:         "Nimble Innovator",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Nimble Innovator — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
