package effects

// Ogre Resister — Creature — Ogre {2}{R}{R}, 4/3 (Mirrodin Besieged,
// #5576):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see
// ordinary_bear.go). The cost, types and body ride game.Card from Scryfall,
// so the Spec is the key and the name.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d626bd7f-c993-49ff-b05b-262330d8540c",
		Name:         "Ogre Resister",
		Completeness: CompletenessFull,
	})
}
