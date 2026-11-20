package effects

// Golden Bear — Creature — Bear {3}{G}, 4/3 (Portal Second Age,
// #5625):
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
		OracleID:     "6ba7ab42-918e-472a-b364-06073cc354d2",
		Name:         "Golden Bear",
		Completeness: CompletenessFull,
	})
}
