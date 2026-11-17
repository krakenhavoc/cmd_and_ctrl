package effects

// Panther Warriors — Creature — Cat Warrior {4}{G}, 6/3 (Visions, #5631):
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
		OracleID:     "ef7788af-8edc-46df-a5b6-895c734ea423",
		Name:         "Panther Warriors",
		Completeness: CompletenessFull,
	})
}
