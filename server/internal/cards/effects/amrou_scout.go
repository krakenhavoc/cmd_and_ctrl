package effects

// Amrou Scout — Creature — Kithkin Rebel Scout {1}{W}, 2/1 (Time Spiral,
// #5596):
//
//	"{4}, {T}: Search your library for a Rebel permanent card with mana
//	 value 3 or less, put it onto the battlefield, then shuffle."
//
// The Rebel recruiter, built by nemRecruiterRow at mana value 3 (the
// Ramosian recruiters' body). The search names a quality, so it may find
// nothing (CR 701.23b), and the library is shuffled either way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4c80112a-8ff5-4ca4-8737-b8d499d78982",
		Name:         "Amrou Scout",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{nemRecruiterRow("Amrou Scout", "{4}", "a", "Rebel", 3)},
	})
}
