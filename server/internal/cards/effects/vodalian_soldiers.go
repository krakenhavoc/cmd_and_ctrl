package effects

// Vodalian Soldiers — Creature — Merfolk Soldier {1}{U}, 1/2
// (Fallen Empires, #5637):
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
		OracleID:     "0e3ac78c-8a98-4786-b86c-ecfe61e13c83",
		Name:         "Vodalian Soldiers",
		Completeness: CompletenessFull,
	})
}
