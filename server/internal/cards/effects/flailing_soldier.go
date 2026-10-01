package effects

// Flailing Soldier — Creature — Human Soldier {R}, 2/2:
//
//	"{1}: This creature gets +1/+1 until end of turn. Any player may
//	 activate this ability.
//	 {1}: This creature gets -1/-1 until end of turn. Any player may
//	 activate this ability."
//
// The Flailing pair (flailing_helpers.go) and nothing else: any player
// may activate either ability (CR 602.2, 602.1b) and pays its {1} out
// of their own pool (CR 602.1a), and the Soldier itself grows or
// shrinks until end of turn (CR 611.2c, 514.2). A -1/-1 that takes its
// toughness to 0 puts it into its owner's graveyard at the next
// state-based check (CR 704.5f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2c1dd6f4-2db5-449d-9ed8-579d98cfcfb8",
		Name:         "Flailing Soldier",
		Completeness: CompletenessFull,
		Activated:    flailingAbilities("Flailing Soldier"),
	})
}
