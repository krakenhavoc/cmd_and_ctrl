package effects

// Thornhide Wolves — Creature — Wolf {4}{G}, 4/5
// (Shadows over Innistrad, #5543):
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
		OracleID:     "c5c221e6-dafe-4661-b920-3faed9551802",
		Name:         "Thornhide Wolves",
		Completeness: CompletenessFull,
	})
}
