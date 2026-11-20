package effects

// Zombie Outlander — Creature — Zombie Scout {U}{B}, 2/2 (Conflux,
// #5587):
//
//	"Protection from green"
//
// Protection from green is a printed keyword, read by targeting,
// blocking, damage and attaching (CR 702.16a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "857524d5-9b5c-42ac-a6cf-2d9bfebd7015",
		Name:            "Zombie Outlander",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"protection from green"},
	})
}
