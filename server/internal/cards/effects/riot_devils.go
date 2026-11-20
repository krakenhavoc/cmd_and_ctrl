package effects

// Riot Devils — Creature — Devil {2}{R}, 2/3 (Innistrad, #5572):
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
		OracleID:     "228ee6b0-2012-4ef8-8aab-552feece801d",
		Name:         "Riot Devils",
		Completeness: CompletenessFull,
	})
}
