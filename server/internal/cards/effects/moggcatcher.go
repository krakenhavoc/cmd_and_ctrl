package effects

// Moggcatcher — Creature — Human Mercenary {2}{R}{R}, 2/2
// (Nemesis, #5618):
//
//	"{3}, {T}: Search your library for a Goblin permanent card, put it onto the battlefield, then shuffle."
//
// The recruiter row (nemRecruiterRow) for a Goblin permanent card, with
// no mana value bound. The searcher may find nothing (CR 701.23b), and
// the library is shuffled either way. The {T} is a creature's, so it
// waits out summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4845dc21-7534-4896-b298-4d44cd7a3b27",
		Name:         "Moggcatcher",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			nemRecruiterRow("Moggcatcher", "{3}", "a", "Goblin", -1),
		},
	})
}
