package effects

// Feral Shadow — Creature — Nightstalker {2}{B}, 2/1 (Portal, #5630):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data (CR 702.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cf26dcb4-e181-4ba5-bc03-b56f57032b85",
		Name:            "Feral Shadow",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
