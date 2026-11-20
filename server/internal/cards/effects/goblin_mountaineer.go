package effects

// Goblin Mountaineer — Creature — Goblin Scout {R}, 1/1 (Portal Second Age,
// #5625):
//
//	"Mountainwalk (This creature can't be blocked as long as defending player controls a Mountain.)"
//
// Mountainwalk is an engine keyword carried as printed data: it can't be
// blocked while the defending player controls a Mountain (CR 702.14c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "356e2df2-2b41-4d9d-8723-3e6a17a090e8",
		Name:            "Goblin Mountaineer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"mountainwalk"},
	})
}
