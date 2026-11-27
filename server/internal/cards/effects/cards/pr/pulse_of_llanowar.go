package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pulse of Llanowar — Enchantment {3}{G} (Invasion):
//
//	"If a basic land you control is tapped for mana, it produces mana of a
//	 color of your choice instead of any other type."
//
// #6210's replacement (TappedForManaProducesAColorOfYourChoice) with
// BasicLandYouControl: the colour is chosen as part of the tap (a mana
// ability can't pause, CR 605.3b), by hand through the five-colour pick and
// by the auto-tapper as the cost needs. The amount of mana is unchanged.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dd1ef824-4e61-4dee-a8f0-9cf459008f2c",
		Name:         "Pulse of Llanowar",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.TappedForManaProducesAColorOfYourChoice(
			"Pulse of Llanowar — if a basic land you control is tapped for mana, it produces mana of a color of your choice instead of any other type.",
			false, effects.BasicLandYouControl)},
	})
}
