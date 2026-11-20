package effects

// Wall of Mist — Creature — Wall {1}{U}, 0/5 (Core Set 2019, #5527):
//
//	"Defender"
//
// The whole text is printed keywords; the engine reads them off
// PrintedKeywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b2ca8824-6bd8-4aa3-b156-618118eb98a4",
		Name:            "Wall of Mist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
	})
}
