package effects

// Yargle, Goliath of Otaria — Legendary Creature — Frog Spirit {4}{U},
// 3/9. No rules text.
//
// A vanilla creature needs no catalog entry for the game to play it; the
// entry exists so the catalogue page lists it as automated rather than
// unaudited. No simplification.
func init() {
	Register(Spec{
		OracleID:     "014d0bcd-b80f-4c43-8f00-f4c31e8f3370",
		Name:         "Yargle, Goliath of Otaria",
		Completeness: CompletenessFull,
	})
}
