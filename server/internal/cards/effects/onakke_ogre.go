package effects

// Onakke Ogre — Creature — Ogre Warrior {2}{R}, 4/2 (Core Set 2019, #5527):
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
		OracleID:     "fa10cacb-ab14-447a-b411-9d74bc3772fb",
		Name:         "Onakke Ogre",
		Completeness: CompletenessFull,
	})
}
