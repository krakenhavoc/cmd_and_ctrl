package effects

// Koskun Keep — Land (Homelands, #5635):
//
//	"{T}: Add {C}.
//	 {1}, {T}: Add {R}.
//	 {2}, {T}: Add {B} or {G}."
//
// Three mana abilities (CR 605.1a), built by tieredManaLand
// (tiered_mana_lands.go) with the rest of its cycle.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "184c5a0d-7654-4421-86e4-7f04bcf49494",
		Name:          "Koskun Keep",
		Completeness:  CompletenessFull,
		ManaAbilities: tieredManaLand("R", "B", "G"),
	})
}
