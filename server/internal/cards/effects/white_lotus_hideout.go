package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// White Lotus Hideout — Land (Avatar: The Last Airbender, #4459):
//
//	"{T}: Add {C}.
//	 {T}: Add one mana of any color. Spend this mana only to cast a Lesson
//	 or Shrine spell.
//	 {1}, {T}: Add one mana of any color."
//
// Three mana abilities. The restricted one is Avengers Tower's
// ManaRestrictAnyOf tag with two spell legs and no ability leg: a spell
// with the subtype Lesson or Shrine may spend it, and nothing else may,
// including an ability of a Shrine (the ruling). The filter is Crystal
// Grotto's {1}, {T} row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2cdbfda3-98fc-4108-b551-c7049168924e",
		Name:         "White Lotus Hideout",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			colorlessTapRow(),
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{W|U|B|R|G}",
				Label:    "{T}: Add one mana of any color (Lesson or Shrine spells only).",
				Restrictions: []string{game.ManaRestrictAnyOf(
					[]string{game.ManaRestrictCast, game.ManaRestrictSubtype("Lesson")},
					[]string{game.ManaRestrictCast, game.ManaRestrictSubtype("Shrine")},
				)},
			},
			{
				Cost:     ManaAbilityCost{Tap: true, Mana: "{1}"},
				Produced: "{W|U|B|R|G}",
				Label:    "{1}, {T}: Add one mana of any color",
			},
		},
	})
}
