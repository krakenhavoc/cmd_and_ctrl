package effects

// Hyena Pack — Creature — Hyena {2}{R}{R}, 3/4
// (Amonkhet, #5535):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see
// grizzly_bears.go). The cost, types and body ride game.Card from
// Scryfall, so the Spec is the key and the name.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d38b027d-eb24-46ba-9d46-3be08eedbe2a",
		Name:         "Hyena Pack",
		Completeness: CompletenessFull,
	})
}
