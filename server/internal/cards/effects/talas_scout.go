package effects

// Talas Scout — Creature — Human Pirate Scout {1}{U}, 1/2 (Portal Second Age,
// #5625):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data: it can be blocked
// only by creatures with flying or reach (CR 702.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "71b7d5c4-c16c-4c6b-8e2e-e9b5cc128e1f",
		Name:            "Talas Scout",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
