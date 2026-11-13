package effects

// Skyshroud Poacher — Creature — Human Rebel {2}{G}{G}, 2/2
// (Nemesis, #5618):
//
//	"{3}, {T}: Search your library for an Elf permanent card, put it onto the battlefield, then shuffle."
//
// The recruiter row (nemRecruiterRow) for an Elf permanent card, with
// no mana value bound. The searcher may find nothing (CR 701.23b), and
// the library is shuffled either way. The {T} is a creature's, so it
// waits out summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4f5921c1-b932-4d6b-bb8e-01992578abdc",
		Name:         "Skyshroud Poacher",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			nemRecruiterRow("Skyshroud Poacher", "{3}", "an", "Elf", -1),
		},
	})
}
