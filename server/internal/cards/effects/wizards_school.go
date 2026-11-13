package effects

// Wizards' School — Land (Homelands, #5635):
//
//	"{T}: Add {C}.
//	 {1}, {T}: Add {U}.
//	 {2}, {T}: Add {W} or {B}."
//
// Three mana abilities (CR 605.1a), built by tieredManaLand
// (tiered_mana_lands.go) with the rest of its cycle.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "51635f96-af1d-4b33-9118-432495aaf07f",
		Name:          "Wizards' School",
		Completeness:  CompletenessFull,
		ManaAbilities: tieredManaLand("U", "W", "B"),
	})
}
