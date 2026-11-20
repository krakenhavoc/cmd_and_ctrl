package effects

// Nettle Swine — {3}{G} Creature — Boar 4/3 (Avacyn Restored, #5570):
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
		OracleID:     "fe412468-b506-4ed8-8fff-3987ef3dc2ee",
		Name:         "Nettle Swine",
		Completeness: CompletenessFull,
	})
}
