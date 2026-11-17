package effects

// Obsianus Golem — Artifact Creature — Golem {6}, 4/6
// (Limited Edition Alpha, #5642):
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
		OracleID:     "ac41171e-c454-49e9-9004-c082ae099630",
		Name:         "Obsianus Golem",
		Completeness: CompletenessFull,
	})
}
