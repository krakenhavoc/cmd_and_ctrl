package effects

// Zephyr Falcon — Creature — Bird {1}{U}, 1/1 (Legends, #5639):
//
//	"Flying, vigilance"
//
// Both are engine keywords: only a flyer or a creature with reach blocks
// it (CR 702.9b), and attacking does not tap it (CR 702.20b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b5e44bd6-0b67-4770-868d-c46af5b6a42f",
		Name:            "Zephyr Falcon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
	})
}
