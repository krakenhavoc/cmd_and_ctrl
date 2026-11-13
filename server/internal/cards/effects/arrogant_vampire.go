package effects

// Arrogant Vampire — Creature — Vampire {3}{B}{B}, 4/3 (Portal, #5630):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data: it can be blocked
// only by creatures with flying or reach (CR 702.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5edfe083-391b-41ae-ac4d-d648042cfa5e",
		Name:            "Arrogant Vampire",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
