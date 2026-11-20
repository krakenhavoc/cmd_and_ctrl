package effects

// Abbey Griffin — Creature — Griffin {3}{W}, 2/2 (Innistrad, #5572):
//
//	"Flying, vigilance"
//
// Flying and vigilance are engine keywords carried as printed data:
// only a creature with flying or reach blocks it (CR 702.9b), and
// attacking doesn't tap it (CR 702.20b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6eb74f6e-1c9d-4caf-a0d8-fa472d435346",
		Name:            "Abbey Griffin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
	})
}
