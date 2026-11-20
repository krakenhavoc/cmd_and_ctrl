package effects

// Skyhunter Patrol — Creature — Cat Knight {2}{W}{W}, 2/3 (Mirrodin, #5606):
//
//	"Flying, first strike"
//
// Two printed keywords the engine already honours: flying (CR 702.9b)
// and first strike (CR 702.7b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "aadcff6f-9207-4d90-a12d-4913c96867e2",
		Name:            "Skyhunter Patrol",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
	})
}
