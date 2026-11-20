package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sylvan Brushstrider — Creature — Beast {2}{G}, 3/2 (Ravnica Allegiance,
// #5521):
//
//	"When this creature enters, you gain 2 life."
//
// Venerable Monk's enters trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "67a494dc-deac-43ee-9e47-20fae9788a9f",
		Name:         "Sylvan Brushstrider",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sylvan Brushstrider — you gain 2 life", Do(GainLife{Amount: 2})),
		},
	})
}
