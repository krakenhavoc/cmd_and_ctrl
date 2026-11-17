package effects

// Glacial Wall — Creature — Wall {2}{U}, 0/7 (Ice Age, #5636):
//
//	"Defender (This creature can't attack.)"
//
// Defender is a printed keyword: it can't attack (CR 702.3b), and it
// blocks as usual.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7fcf8faa-2c93-4dc8-aa56-6fc40340e212",
		Name:            "Glacial Wall",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
	})
}
