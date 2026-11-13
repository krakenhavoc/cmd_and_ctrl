package effects

// The Lady of the Mountain — Legendary Creature — Giant {4}{R}{G}, 5/5 (Legends, #5639)
//
// No rules text: a vanilla 5/5. Registered so the catalogue lists it
// as reviewed; there is nothing for the engine to do beyond its printed
// power and toughness.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3ef03975-0f46-43aa-b461-bd3b6f5d4613",
		Name:         "The Lady of the Mountain",
		Completeness: CompletenessFull,
	})
}
