package effects

// Canopy Gorger — Creature — Wurm {4}{G}{G}, 6/5 (Oath of the Gatewatch, #5545):
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
		OracleID:     "1d614b37-d578-4f35-8fd8-3c7800d5f742",
		Name:         "Canopy Gorger",
		Completeness: CompletenessFull,
	})
}
