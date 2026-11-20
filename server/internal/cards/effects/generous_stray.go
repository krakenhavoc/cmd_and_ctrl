package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Generous Stray — Creature — Cat {2}{G}, 1/2 (Guilds of Ravnica, #5525):
//
//	"When this creature enters, draw a card."
//
// Striped Bears' enter trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5722a13f-7d70-438c-94ba-97a3f53fbc5c",
		Name:         "Generous Stray",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Generous Stray — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
