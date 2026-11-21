package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Honey Mammoth — Creature — Elephant {4}{G}{G}, 6/6 (Ikoria: Lair of Behemoths, #5512):
//
//	"When this creature enters, you gain 4 life."
//
// Teroh's Faithful's entry trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "41da0e8c-ba86-446d-9719-667695c7f1cc",
		Name:         "Honey Mammoth",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{LifeGain: 4},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Honey Mammoth — you gain 4 life", effects.Do(effects.GainLife{Amount: 4})),
		},
	})
}
