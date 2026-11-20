package effects

// Exultant Skymarcher — Creature — Vampire Soldier {1}{W}{W}, 2/3 (Rivals of Ixalan, #5531):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data: only a creature
// with flying or reach can block it (CR 702.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d7ebd4a7-c542-469f-8a1c-d48b92fa540a",
		Name:            "Exultant Skymarcher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
