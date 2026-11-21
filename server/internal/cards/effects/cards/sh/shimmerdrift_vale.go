package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shimmerdrift Vale — Snow Land (Jumpstart 2022, #5948):
//
//	"This land enters tapped.
//	 As this land enters, choose a color.
//	 {T}: Add one mana of the chosen color."
//
// Edgewall Inn's first three lines: the tapped entry, the colour prompt
// as it enters, and mana of the chosen colour.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2ac34f3e-822d-4fde-99ca-a31c4d9503fd",
		Name:         "Shimmerdrift Vale",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		AsEnters:     effects.ChooseColorAsEnters(game.ColorForMana, "Shimmerdrift Vale"),
		ManaAbilities: []effects.ManaAbility{{
			Cost:         effects.ManaAbilityCost{Tap: true},
			ProducedFunc: effects.ProducedChosenColor(),
			Label:        "Add one mana of the chosen color",
		}},
	})
}
