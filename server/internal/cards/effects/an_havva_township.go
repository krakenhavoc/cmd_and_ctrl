package effects

// An-Havva Township — Land (Homelands, #5635):
//
//	"{T}: Add {C}.
//	 {1}, {T}: Add {G}.
//	 {2}, {T}: Add {R} or {W}."
//
// Three mana abilities (CR 605.1a), built by tieredManaLand
// (tiered_mana_lands.go) with the rest of its cycle.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "40ae17be-9998-4ee4-9d95-82a08895405f",
		Name:          "An-Havva Township",
		Completeness:  CompletenessFull,
		ManaAbilities: tieredManaLand("G", "R", "W"),
	})
}
