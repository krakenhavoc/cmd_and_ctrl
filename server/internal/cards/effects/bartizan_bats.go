package effects

// Bartizan Bats — Creature — Bat {3}{B}, 3/1 (Guilds of Ravnica, #5525):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data (CR 702.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6987d5d4-0ef6-4326-a590-678ac160fe56",
		Name:            "Bartizan Bats",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
