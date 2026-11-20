package effects

// Eager Cadet — Creature — Human Soldier {W}, 1/1
// (Starter 1999, #5620):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (Grizzly Bears'
// shape). The cost, types and body ride game.Card from Scryfall, so the
// Spec is the key and the name.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "39780477-01ba-40c6-8fe6-2178efb10b92",
		Name:         "Eager Cadet",
		Completeness: CompletenessFull,
	})
}
