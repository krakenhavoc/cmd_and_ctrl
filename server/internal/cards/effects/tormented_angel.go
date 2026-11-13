package effects

// Tormented Angel — Creature — Angel {3}{W}, 1/5 (Urza's Destiny, #5621):
//
//	"Flying"
//
// One printed keyword, which the block check reads: only creatures with
// flying or reach can block it (CR 702.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "176eeafd-0482-4fa4-9174-10f8b9494d9b",
		Name:            "Tormented Angel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
