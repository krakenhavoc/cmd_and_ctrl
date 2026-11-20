package effects

// Border Guard — Creature — Human Soldier {2}{W}, 1/4 (Portal, #5630):
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
		OracleID:     "1ef5003c-f540-4cdc-913f-7d5280ad9f62",
		Name:         "Border Guard",
		Completeness: CompletenessFull,
	})
}
