package effects

// Thraben Valiant — {1}{W} Creature — Human Soldier 2/1 (Avacyn Restored, #5570):
//
//	"Vigilance"
//
// One printed keyword (CR 702.20).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "dee0b358-b9da-49b9-a47f-7907cde6d8ee",
		Name:            "Thraben Valiant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
}
