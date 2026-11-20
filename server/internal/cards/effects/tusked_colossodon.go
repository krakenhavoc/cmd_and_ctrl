package effects

// Tusked Colossodon — Creature — Beast {4}{G}{G}, 6/5 (Khans of Tarkir,
// #5554):
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
		OracleID:     "bba11c0e-01a3-45ad-8de4-2148cf660ba8",
		Name:         "Tusked Colossodon",
		Completeness: CompletenessFull,
	})
}
