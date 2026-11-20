package effects

// Wasp Lancer — Creature — Faerie Soldier {U/B}{U/B}{U/B}, 3/2
// (Shadowmoor, #5591):
//
//	"Flying"
//
// Flying is a printed keyword (CR 702.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "12a9a9a4-6b8c-48a1-833f-457a7ece8e7e",
		Name:            "Wasp Lancer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
