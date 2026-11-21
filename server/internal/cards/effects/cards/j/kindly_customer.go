package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kindly Customer — Creature — Human Citizen {1}{W}, 1/1 (Avatar: The
// Last Airbender Eternal, #5075):
//
//	"When this creature enters, draw a card."
//
// Elvish Visionary's enters trigger: it goes on the stack and the
// controller draws when it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3faab84f-7055-4554-b948-181928d8d4e2",
		Name:         "Kindly Customer",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Kindly Customer — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
