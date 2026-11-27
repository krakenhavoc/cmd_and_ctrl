package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ashling's Prerogative — Enchantment {1}{R} (Lorwyn):
//
//	"As this enchantment enters, choose odd or even. (Zero is even.)
//	 Each creature with mana value of the chosen quality has haste.
//	 Each creature without mana value of the chosen quality enters tapped."
//
// The Siege pattern (CR 614.12): the controller picks Odd or Even as it
// enters, onto Card.ChosenOption, and both lines exist only while a word is
// chosen. Haste is a layer-6 grant to every creature, any player's, whose
// mana value has the chosen parity; the tapped entry is a CR 614 replacement
// on the move of every creature of the other parity. A token or a face-down
// creature has mana value 0, which is even.
//
// No simplification.
func init() {
	var statics []game.StaticAbility
	var reps []game.ReplacementEffect
	for _, option := range []string{"Odd", "Even"} {
		haste := effects.KeywordGrant(effects.S82OldB550ManaValueParity(option == "Odd"), "haste")
		haste.Label = "Each creature with mana value of the chosen quality has haste."
		statics = append(statics, effects.StaticWhenChosen(option, haste))
		tapped := effects.S82OldB550EntersTappedUnlessParity(option == "Odd", "Ashling's Prerogative: enters tapped")
		tapped.ActiveWhen = effects.ChosenIs(option)
		reps = append(reps, tapped)
	}
	effects.Register(effects.Spec{
		OracleID:     "b68bc097-482a-4fe5-9fca-df5941f3dc5c",
		Name:         "Ashling's Prerogative",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseOptionAsEnters("Ashling's Prerogative", "Odd", "Even"),
		Static:       statics,
		Replacements: reps,
	})
}
