package effects

// Blood Vassal — Creature — Thrull {2}{B}, 2/2 (Urza's Saga):
//
//	"Sacrifice this creature: Add {B}{B}."
//
// Blood Pet's mana ability at two black: a mana ability (CR 605.1a), so
// it doesn't use the stack, and with no {T} in its cost it can be
// activated the turn the Thrull arrives.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f792f40d-1d37-4854-98da-e4020c6a44d4",
		Name:         "Blood Vassal",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Sacrifice: true},
			Produced: "{B}{B}",
			Label:    "Sacrifice this creature: Add {B}{B}",
		}},
	})
}
