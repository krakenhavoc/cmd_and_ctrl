package effects

// Blood Celebrant — Creature — Human Cleric {B}, 1/1 (Legions, #5609):
//
//	"{B}, Pay 1 life: Add one mana of any color."
//
// A mana ability (CR 605.1a) with a mana and a life component in its
// cost: the {B} is paid from the pool and the life is checked before
// anything is paid (CR 119.4), so at 0 life it can't be activated. There
// is no {T}, so it works the turn the Celebrant arrives and any number
// of times.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "695ce90b-a5a2-4d8c-ac05-07f04010603e",
		Name:         "Blood Celebrant",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Mana: "{B}", Life: 1},
			Produced: "{W|U|B|R|G}",
			Label:    "{B}, Pay 1 life: Add one mana of any color.",
		}},
	})
}
