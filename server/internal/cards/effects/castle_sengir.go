package effects

// Castle Sengir — Land (Homelands, #5635):
//
//	"{T}: Add {C}.
//	 {1}, {T}: Add {B}.
//	 {2}, {T}: Add {U} or {R}."
//
// Three mana abilities (CR 605.1a), built by tieredManaLand
// (tiered_mana_lands.go) with the rest of its cycle.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "c7f0251a-9341-4ff2-8b15-31c06eb4f2e7",
		Name:          "Castle Sengir",
		Completeness:  CompletenessFull,
		ManaAbilities: tieredManaLand("B", "U", "R"),
	})
}
