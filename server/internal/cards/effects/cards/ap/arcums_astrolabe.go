package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arcum's Astrolabe — Snow Artifact {S} (Modern Horizons):
//
//	"({S} can be paid with one mana from a snow source.)
//	 When this artifact enters, draw a card.
//	 {1}, {T}: Add one mana of any color."
//
// The {S} is in the mana cost as printed (CR 107.4h): only mana a snow source
// made pays it. The enters trigger draws a card. The mana ability (CR 605.1a)
// is Abstergo Entertainment's, {1} and tap for one mana of any color.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ec463bde-dadf-4044-8e41-71338fd4d62f",
		Name:         "Arcum's Astrolabe",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Arcum's Astrolabe — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true, Mana: "{1}"},
			Produced: "{W|U|B|R|G}",
			Label:    "{1}, {T}: Add one mana of any color.",
		}},
	})
}
