package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sylvan Brushstrider — Creature — Beast {2}{G}, 3/2 (Ravnica Allegiance,
// #5521):
//
//	"When this creature enters, you gain 2 life."
//
// Venerable Monk's enters trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "67a494dc-deac-43ee-9e47-20fae9788a9f",
		Name:         "Sylvan Brushstrider",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sylvan Brushstrider — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}
