package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Generous Stray — Creature — Cat {2}{G}, 1/2 (Guilds of Ravnica, #5525):
//
//	"When this creature enters, draw a card."
//
// Striped Bears' enter trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5722a13f-7d70-438c-94ba-97a3f53fbc5c",
		Name:         "Generous Stray",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Generous Stray — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
