package effects

// Flailing Ogre — Creature — Ogre {2}{R}, 3/3:
//
//	"{1}: This creature gets +1/+1 until end of turn. Any player may
//	 activate this ability.
//	 {1}: This creature gets -1/-1 until end of turn. Any player may
//	 activate this ability."
//
// The Flailing pair (flailing_helpers.go) and nothing else: any player
// may activate either ability (CR 602.2, 602.1b) and pays its {1} out
// of their own pool (CR 602.1a), and the Ogre itself grows or shrinks
// until end of turn (CR 611.2c, 514.2). A -1/-1 that takes its
// toughness to 0 puts it into its owner's graveyard at the next
// state-based check (CR 704.5f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e81fe764-8b26-40b3-8e17-741ab5e231b4",
		Name:         "Flailing Ogre",
		Completeness: CompletenessFull,
		Activated:    flailingAbilities("Flailing Ogre"),
	})
}
