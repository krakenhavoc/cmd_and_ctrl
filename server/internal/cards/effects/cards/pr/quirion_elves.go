package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quirion Elves — Creature — Elf Druid {1}{G}, 1/1 (Invasion, #5616):
//
//	"As this creature enters, choose a color.
//	 {T}: Add {G}.
//	 {T}: Add one mana of the chosen color."
//
// The colour prompt is the #742 one on AsEnters (Coldsteel Heart's); the two
// mana abilities are separate printed abilities. Until a colour is chosen the
// second produces nothing, the weaker direction and unobservable because the
// open prompt holds priority.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e2975a88-5a5a-462a-90b9-d8c00150b6d3",
		Name:         "Quirion Elves",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseColorAsEnters(game.ColorForMana, "Quirion Elves"),
		ManaAbilities: []effects.ManaAbility{
			{Cost: effects.ManaAbilityCost{Tap: true}, Produced: "{G}", Label: "{T}: Add {G}"},
			{
				Cost:         effects.ManaAbilityCost{Tap: true},
				ProducedFunc: effects.ProducedChosenColor(),
				Label:        "{T}: Add one mana of the chosen color",
			},
		},
	})
}
