package effects

// Wall of Granite — Creature — Wall {2}{R}, 0/7 (Portal, #5630):
//
//	"Defender"
//
// Defender is an engine keyword carried as printed data: it can't
// attack (CR 702.3b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8445094f-008b-491a-977c-e8582d5ab72c",
		Name:            "Wall of Granite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
	})
}
