package effects

// Nyxborn Brute — Enchantment Creature — Cyclops {3}{R}{R}, 7/3 (Theros Beyond Death, #5514):
//
//	A vanilla creature: no rules text.
//
// Nothing to implement: the type line and the printed power and
// toughness come from the card data, as for Jasmine Boreal.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "986ba716-3921-4f64-9acf-703200d666f6",
		Name:         "Nyxborn Brute",
		Completeness: CompletenessFull,
	})
}
