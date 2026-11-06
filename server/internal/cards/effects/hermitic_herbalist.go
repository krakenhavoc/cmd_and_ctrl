package effects

// Hermitic Herbalist — Creature — Human Druid Ally {G}{U}, 2/3 (Avatar:
// The Last Airbender, #4459):
//
//	"{T}: Add one mana of any color.
//	 {T}: Add two mana in any combination of colors. Spend this mana only
//	 to cast Lesson spells."
//
// Two mana abilities. The second's mana carries a spend restriction
// (CR 106.6): only a spell being cast with the Lesson subtype may use it,
// and no activated ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9928773d-3f35-4b0e-965d-4233e04398fb",
		Name:         "Hermitic Herbalist",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{W|U|B|R|G}",
				Label:    "Add one mana of any color",
			},
			{
				Cost:         ManaAbilityCost{Tap: true},
				Produced:     AnyCombinationOfColors(2),
				Restrictions: []string{ManaRestrictCast, ManaRestrictSubtype("Lesson")},
				Label:        "Add two mana in any combination of colors. Spend this mana only to cast Lesson spells",
			},
		},
	})
}
