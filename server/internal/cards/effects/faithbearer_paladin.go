package effects

// Faithbearer Paladin — Creature — Human Knight {4}{W}, 3/4 (Eldritch
// Moon, #5541):
//
//	"Lifelink"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "725dc451-cae4-4345-9741-6c58e5a1a1dc",
		Name:            "Faithbearer Paladin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
	})
}
