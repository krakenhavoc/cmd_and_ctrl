package effects

// Seahunter — Creature — Human Mercenary {2}{U}{U}, 2/2
// (Nemesis, #5618):
//
//	"{3}, {T}: Search your library for a Merfolk permanent card, put it onto the battlefield, then shuffle."
//
// The recruiter row (nemRecruiterRow) for a Merfolk permanent card,
// with no mana value bound. The searcher may find nothing (CR 701.23b),
// and the library is shuffled either way. The {T} is a creature's, so
// it waits out summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "59e1899f-a9e8-48aa-b3fb-0b7d9fd6859c",
		Name:         "Seahunter",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			nemRecruiterRow("Seahunter", "{3}", "a", "Merfolk", -1),
		},
	})
}
