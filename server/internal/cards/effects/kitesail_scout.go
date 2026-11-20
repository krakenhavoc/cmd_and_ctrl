package effects

// Kitesail Scout — Creature — Kor Scout {W}, 1/1 (Battle for Zendikar,
// #5547):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data: it can be
// blocked only by creatures with flying or reach (CR 702.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b365a7dd-fed1-40f8-9544-2b5bc1af761c",
		Name:            "Kitesail Scout",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
