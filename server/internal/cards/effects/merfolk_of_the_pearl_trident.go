package effects

// Merfolk of the Pearl Trident — Creature — Merfolk {U}, 1/1
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
		OracleID:     "218d9277-c179-4de3-9c7f-79b5a6d4fa38",
		Name:         "Merfolk of the Pearl Trident",
		Completeness: CompletenessFull,
	})
}
