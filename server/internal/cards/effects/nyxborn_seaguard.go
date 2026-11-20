package effects

// Nyxborn Seaguard — Enchantment Creature — Merfolk Soldier {2}{U}{U}, 2/5 (Theros Beyond Death, #5514):
//
//	A vanilla creature: no rules text.
//
// Nothing to implement: the type line and the printed power and
// toughness come from the card data, as for Jasmine Boreal.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "48cc79b8-f7a5-474b-bd88-7cd6ae673086",
		Name:         "Nyxborn Seaguard",
		Completeness: CompletenessFull,
	})
}
