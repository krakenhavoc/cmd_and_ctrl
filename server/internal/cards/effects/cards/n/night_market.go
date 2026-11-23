package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Night Market — Land (Aetherdrift, #4464):
//
//	"This land enters tapped. As it enters, choose a color.
//	 {T}: Add one mana of the chosen color.
//	 Cycling {3} ({3}, Discard this card: Draw a card.)"
//
// Uncharted Haven (chosen_color_lands.go) with cycling: the tapped entry
// and the colour prompt are both replacements as it enters (CR 614.1c),
// and the mana ability reads the stored colour when it is activated.
// Cycling is the hand ability (CR 702.29a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4cdb9f80-d08f-4986-99c8-573166d66082",
		Name:         "Night Market",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		AsEnters:     effects.ChooseColorAsEnters(game.ColorForMana, "Night Market"),
		ManaAbilities: []effects.ManaAbility{{
			Cost:         effects.ManaAbilityCost{Tap: true},
			ProducedFunc: effects.ProducedChosenColor(),
			Label:        "Add one mana of the chosen color",
		}},
		Activated: []effects.ActivatedAbility{effects.Cycling("{3}")},
	})
}
