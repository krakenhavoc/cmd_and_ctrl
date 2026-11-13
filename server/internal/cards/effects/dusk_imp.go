package effects

// Dusk Imp — Creature — Imp {2}{B}, 2/1 (Odyssey, #5613):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data: it can be blocked
// only by creatures with flying or reach (CR 702.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1389ae4a-c3a5-4678-9012-937a3cbaf7f7",
		Name:            "Dusk Imp",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
