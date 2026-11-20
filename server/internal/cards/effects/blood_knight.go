package effects

// Blood Knight — Creature — Human Knight {R}{R}, 2/2 (Planar Chaos, #5595):
//
//	"First strike, protection from white"
//
// Both are engine keywords carried as printed data: first strike (CR
// 702.7b) and protection from white (CR 702.16b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "67fba605-9cfa-499c-83e0-4fbd023bcfa0",
		Name:            "Blood Knight",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "protection from white"},
	})
}
