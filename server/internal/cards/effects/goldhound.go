package effects

// Goldhound — Artifact Creature — Treasure Dog {R}, 1/1 (Streets of New
// Capenna, #5491):
//
//	"First strike
//	 Menace
//	 {T}, Sacrifice this creature: Add one mana of any color."
//
// Buried Treasure's mana ability on a creature: a mana ability (CR 605.1a)
// whose cost taps and sacrifices the source.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b5eb72bb-897f-4911-a797-3153c31df3e0",
		Name:            "Goldhound",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "menace"},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true, Sacrifice: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}, Sacrifice this creature: Add one mana of any color",
		}},
	})
}
