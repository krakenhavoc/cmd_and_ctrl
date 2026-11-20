package effects

// Goblin Outlander — Creature — Goblin Scout {B}{R}, 2/2 (Conflux,
// #5587):
//
//	"Protection from white"
//
// Protection from white is a printed keyword, read by targeting,
// blocking, damage and attaching (CR 702.16a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3cc99034-6eaf-4e45-9bff-57b74847343e",
		Name:            "Goblin Outlander",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"protection from white"},
	})
}
