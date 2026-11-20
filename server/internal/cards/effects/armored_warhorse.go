package effects

// Armored Warhorse — Creature — Horse {W}{W}, 2/3 (Magic 2012, #5573):
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
		OracleID:     "29233aa4-3fa0-4074-ad0f-c8e7101052f3",
		Name:         "Armored Warhorse",
		Completeness: CompletenessFull,
	})
}
