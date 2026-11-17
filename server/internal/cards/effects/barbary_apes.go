package effects

// Barbary Apes — Creature — Ape {1}{G}, 2/2 (Legends, #5639)
//
// No rules text: a vanilla 2/2. Registered so the catalogue lists it
// as reviewed; there is nothing for the engine to do beyond its printed
// power and toughness.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "39bfe194-1662-4c0e-a7c6-ea7764d29b65",
		Name:         "Barbary Apes",
		Completeness: CompletenessFull,
	})
}
