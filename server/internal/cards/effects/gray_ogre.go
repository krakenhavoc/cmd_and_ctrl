package effects

// Gray Ogre — Creature — Ogre {2}{R}, 2/2
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
		OracleID:     "83c8a3a6-2e1a-4e26-8847-6d066f42d906",
		Name:         "Gray Ogre",
		Completeness: CompletenessFull,
	})
}
