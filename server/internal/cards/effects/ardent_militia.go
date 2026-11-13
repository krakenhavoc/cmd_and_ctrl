package effects

// Ardent Militia — Creature — Human Soldier {4}{W}, 2/5 (Portal, #5630):
//
//	"Vigilance"
//
// Vigilance is an engine keyword carried as printed data: attacking
// doesn't tap it (CR 702.20b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "23625877-b6db-480c-8885-a62b7d0457df",
		Name:            "Ardent Militia",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
}
