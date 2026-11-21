package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "26e4b49e-77e7-41d9-94c5-924669a82591",
		Name:         "Gallifrey Council Chamber",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Gallifrey Council Chamber — surveil 1", Do(Surveil{N: 1})),
		},
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "{T}: Add {C}.",
			},
			{
				Cost:     ManaAbilityCost{Tap: true},
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
