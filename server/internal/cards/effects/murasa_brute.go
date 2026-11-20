package effects

// Murasa Brute — Creature — Troll Warrior {2}{G}, 3/3 (Zendikar Rising,
// #5508):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see
// ordinary_bear.go). The cost, types and body ride game.Card from
// Scryfall, so the Spec is the key and the name.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5f84a140-577f-4a1e-a900-793664332fd6",
		Name:         "Murasa Brute",
		Completeness: CompletenessFull,
	})
}
