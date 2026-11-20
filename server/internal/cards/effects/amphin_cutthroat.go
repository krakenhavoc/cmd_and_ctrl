package effects

// Amphin Cutthroat — Creature — Salamander Rogue {3}{U}, 2/4 (Magic 2012,
// #5573):
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
		OracleID:     "e766a5eb-684b-4939-b164-6093d15600c9",
		Name:         "Amphin Cutthroat",
		Completeness: CompletenessFull,
	})
}
