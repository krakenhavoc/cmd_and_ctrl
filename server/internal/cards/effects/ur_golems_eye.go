package effects

// Ur-Golem's Eye — Artifact {4} (Darksteel, #5605):
//
//	"{T}: Add {C}{C}."
//
// Sol Ring's mana ability (CR 605.1a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fb34fc00-e60d-41fc-9393-ca4248ec0a1c",
		Name:         "Ur-Golem's Eye",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}{C}",
			Label:    "Add {C}{C}",
		}},
	})
}
