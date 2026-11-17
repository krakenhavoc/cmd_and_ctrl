package effects

// Rathi Intimidator — Creature — Phyrexian Horror Mercenary {1}{B}{B}, 2/1
// (Nemesis, #5618):
//
//	"Fear (This creature can't be blocked except by artifact creatures and/or black creatures.)
//	 {2}, {T}: Search your library for a Mercenary permanent card with mana value 2 or less, put it onto the battlefield, then shuffle."
//
// Fear is printed data. The recruiter row (nemRecruiterRow) for a
// Mercenary permanent card with mana value 2 or less, the Ramosian
// recruiters' body. The searcher may find nothing (CR 701.23b), and the
// library is shuffled either way. The {T} is a creature's, so it waits
// out summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a24f4198-2412-4b73-8e20-4ced3a826301",
		Name:            "Rathi Intimidator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"fear"},
		Activated: []ActivatedAbility{
			nemRecruiterRow("Rathi Intimidator", "{2}", "a", "Mercenary", 2),
		},
	})
}
