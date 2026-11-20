package effects

// Alabaster Kirin — Creature — Kirin {3}{W}, 2/3 (Khans of Tarkir,
// #5554):
//
//	"Flying, vigilance"
//
// Flying and vigilance are engine keywords carried as printed data: it
// can be blocked only by creatures with flying or reach (CR 702.9b), and
// attacking doesn't tap it (CR 702.20b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "640fa591-ba79-4148-86a6-9de994a9e67c",
		Name:            "Alabaster Kirin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
	})
}
