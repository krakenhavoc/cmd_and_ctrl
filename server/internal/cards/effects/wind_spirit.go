package effects

// Wind Spirit — Creature — Elemental Spirit {4}{U}, 3/2 (Ice Age,
// #5636):
//
//	"Flying
//	 Menace (This creature can't be blocked except by two or more
//	 creatures.)"
//
// Both are printed keywords the block legality check reads: only
// creatures with flying or reach can block it (CR 702.9b), and only two
// or more of them together (CR 702.111b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d4c22b68-c4cb-447a-97df-431e2f1e31f2",
		Name:            "Wind Spirit",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "menace"},
	})
}
