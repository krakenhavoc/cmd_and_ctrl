package effects

// Cliffhaven Sell-Sword — Creature — Kor Warrior {1}{W}, 3/1 (Zendikar
// Rising, #5508):
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
		OracleID:     "5f0c8c60-96f1-4b26-a9f6-386e1cec517d",
		Name:         "Cliffhaven Sell-Sword",
		Completeness: CompletenessFull,
	})
}
