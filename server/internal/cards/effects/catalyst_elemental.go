package effects

// Catalyst Elemental — Creature — Elemental {2}{R}, 2/2
// (Core Set 2019, #5527):
//
//	"Sacrifice this creature: Add {R}{R}."
//
// Reckless Barbarian's text: a mana ability whose cost is sacrificing
// the creature (CR 605.1a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7779525d-080d-4669-bb65-af532bf0983e",
		Name:         "Catalyst Elemental",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Sacrifice: true},
			Produced: "{R}{R}",
			Label:    "Sacrifice this creature: Add {R}{R}",
		}},
	})
}
