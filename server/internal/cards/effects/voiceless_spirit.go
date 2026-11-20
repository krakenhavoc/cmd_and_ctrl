package effects

// Voiceless Spirit — Creature — Spirit {2}{W}, 2/1 (Innistrad, #5572):
//
//	"Flying, first strike"
//
// Flying and first strike are engine keywords carried as printed data
// (CR 702.9b, CR 702.7b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "56cb7592-ad58-42dc-b568-4b7754bba53d",
		Name:            "Voiceless Spirit",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
	})
}
