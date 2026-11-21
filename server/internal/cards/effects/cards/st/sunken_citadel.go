package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sunken Citadel — Land — Cave (The Lost Caverns of Ixalan, #4470):
//
//	"This land enters tapped. As it enters, choose a color.
//	 {T}: Add one mana of the chosen color.
//	 {T}: Add two mana of the chosen color. Spend this mana only to
//	 activate abilities of land sources."
//
// The one-colour land shape (chosen_color_lands.go) with a second mana
// ability: two mana of the chosen colour, restricted to activating
// abilities (CR 106.6) of land sources, read from the source of the ability
// being paid for (a land's mana ability counts, CR 605.1a; a spell, even a
// land-creature's, never). Before a colour is chosen both produce nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "508189e1-9cef-4f9c-8ff1-078c99a0f603",
		Name:         "Sunken Citadel",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		AsEnters:     effects.ChooseColorAsEnters(game.ColorForMana, "Sunken Citadel"),
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:         effects.ManaAbilityCost{Tap: true},
				ProducedFunc: effects.ProducedChosenColor(),
				Label:        "Add one mana of the chosen color",
			},
			{
				Cost:           effects.ManaAbilityCost{Tap: true},
				ProducedAmount: effects.ProducedChosenColorAmount(2),
				Label:          "{T}: Add two mana of the chosen color. Spend this mana only to activate abilities of land sources",
				Restrictions: []string{game.ManaRestrictAnyOf(
					[]string{game.ManaRestrictActivate, game.ManaRestrictType("Land")},
				)},
			},
		},
	})
}
