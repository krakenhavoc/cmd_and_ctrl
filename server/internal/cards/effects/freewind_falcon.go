package effects

// Freewind Falcon — Creature — Bird {1}{W}, 1/1 (Visions, #5631):
//
//	"Flying, protection from red"
//
// Both are engine keywords carried as printed data: flying limits what
// can block it (CR 702.9b), and protection from red stops red spells
// and abilities from red sources targeting it and prevents damage from
// red sources (CR 702.16b, CR 702.16e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8cdda36f-5e82-4c2f-8eaa-75b9bbc93a02",
		Name:            "Freewind Falcon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "protection from red"},
	})
}
