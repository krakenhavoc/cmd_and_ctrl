package effects

// Earth Elemental — Creature — Elemental {3}{R}{R}, 4/5
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
		OracleID:     "3c97c311-7ad5-47ec-b421-f6c3bfbda9fb",
		Name:         "Earth Elemental",
		Completeness: CompletenessFull,
	})
}
