package effects

// Aysen Abbey — Land (Homelands, #5635):
//
//	"{T}: Add {C}.
//	 {1}, {T}: Add {W}.
//	 {2}, {T}: Add {G} or {U}."
//
// Three mana abilities (CR 605.1a), built by tieredManaLand
// (tiered_mana_lands.go) with the rest of its cycle.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "6ff85e73-bf7a-4a9c-80ef-6ce76656fab7",
		Name:          "Aysen Abbey",
		Completeness:  CompletenessFull,
		ManaAbilities: tieredManaLand("W", "G", "U"),
	})
}
