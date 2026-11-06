package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "4cdb9f80-d08f-4986-99c8-573166d66082",
		Name:         "Night Market",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		AsEnters:     ChooseColorAsEnters(game.ColorForMana, "Night Market"),
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			ProducedFunc: ProducedChosenColor(),
			Label:        "Add one mana of the chosen color",
		}},
		Activated: []ActivatedAbility{Cycling("{3}")},
	})
}
