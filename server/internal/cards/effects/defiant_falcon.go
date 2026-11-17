package effects

// Defiant Falcon — Creature — Rebel Bird {1}{W}, 1/1
// (Nemesis, #5618):
//
//	"Flying
//	 {4}, {T}: Search your library for a Rebel permanent card with mana value 3 or less, put it onto the battlefield, then shuffle."
//
// Flying is printed data. The recruiter row (nemRecruiterRow) for a
// Rebel permanent card with mana value 3 or less, the Ramosian
// recruiters' body. The searcher may find nothing (CR 701.23b), and the
// library is shuffled either way. The {T} is a creature's, so it waits
// out summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "dc206e6a-e447-48d4-a91a-7c746320e2b4",
		Name:            "Defiant Falcon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{
			nemRecruiterRow("Defiant Falcon", "{4}", "a", "Rebel", 3),
		},
	})
}
