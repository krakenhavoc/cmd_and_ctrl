package effects

// Forest Bear — Creature — Bear {1}{G}, 2/2 (Portal Three Kingdoms, #5622):
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
		OracleID:     "0994d596-0b15-4156-8e94-5b566dcad6cd",
		Name:         "Forest Bear",
		Completeness: CompletenessFull,
	})
}
