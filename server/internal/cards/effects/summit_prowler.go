package effects

// Summit Prowler — Creature — Yeti {2}{R}{R}, 4/3 (Khans of Tarkir,
// #5554):
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
		OracleID:     "c63a9a3d-13df-4a94-a8e1-18071c351a8d",
		Name:         "Summit Prowler",
		Completeness: CompletenessFull,
	})
}
