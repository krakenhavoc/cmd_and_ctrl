package effects

// Armored Pegasus — Creature — Pegasus {1}{W}, 1/2 (Portal, #5630):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data: it can be blocked
// only by creatures with flying or reach (CR 702.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f097a059-5505-4c3c-b879-7853ab6972ed",
		Name:            "Armored Pegasus",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
