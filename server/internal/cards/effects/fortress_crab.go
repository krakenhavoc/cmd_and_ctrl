package effects

// Fortress Crab — Creature — Crab {3}{U}, 1/6 (Innistrad, #5572):
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
		OracleID:     "075fd933-689c-4199-85bd-e7994d936145",
		Name:         "Fortress Crab",
		Completeness: CompletenessFull,
	})
}
