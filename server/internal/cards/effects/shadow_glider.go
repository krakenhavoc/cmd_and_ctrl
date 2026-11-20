package effects

// Shadow Glider — Creature — Kor Soldier {2}{W}, 2/2 (Battle for
// Zendikar, #5547):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data: it can be
// blocked only by creatures with flying or reach (CR 702.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b92e0156-ea0f-4b13-b3bc-7e84a31445fc",
		Name:            "Shadow Glider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
