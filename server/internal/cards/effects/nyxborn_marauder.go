package effects

// Nyxborn Marauder — Enchantment Creature — Minotaur {2}{B}{B}, 4/3 (Theros Beyond Death, #5514):
//
//	A vanilla creature: no rules text.
//
// Nothing to implement: the type line and the printed power and
// toughness come from the card data, as for Jasmine Boreal.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1552c6d1-f20a-4303-acf2-2fdd432a98e4",
		Name:         "Nyxborn Marauder",
		Completeness: CompletenessFull,
	})
}
