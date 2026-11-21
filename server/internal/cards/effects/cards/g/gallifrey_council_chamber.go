package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gallifrey Council Chamber — Legendary Land (Doctor Who):
//
//	"When Gallifrey Council Chamber enters, surveil 1.
//	 {T}: Add {C}.
//	 {T}: Add one mana of any color. Spend this mana only to cast a Time
//	 Lord or Alien spell or activate an ability of a Time Lord or Alien."
//
// The restricted mana is one anyOf tag over four alternatives: cast or
// activate, each for a source with the Alien subtype or with both type words
// Time and Lord (the engine reads a type line word by word, so Time Lord is
// two subtypes, and no creature type is spelled Time or Lord otherwise).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "26e4b49e-77e7-41d9-94c5-924669a82591",
		Name:         "Gallifrey Council Chamber",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Gallifrey Council Chamber — surveil 1", effects.Do(effects.Surveil{N: 1})),
		},
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "{T}: Add {C}.",
			},
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{W|U|B|R|G}",
				Label:    "{T}: Add one mana of any color. Spend this mana only to cast a Time Lord or Alien spell or activate an ability of a Time Lord or Alien.",
				Restrictions: []string{game.ManaRestrictAnyOf(
					[]string{game.ManaRestrictCast, game.ManaRestrictSubtype("Alien")},
					[]string{game.ManaRestrictCast, game.ManaRestrictSubtype("Time"), game.ManaRestrictSubtype("Lord")},
					[]string{game.ManaRestrictActivate, game.ManaRestrictSubtype("Alien")},
					[]string{game.ManaRestrictActivate, game.ManaRestrictSubtype("Time"), game.ManaRestrictSubtype("Lord")},
				)},
			},
		},
	})
}
