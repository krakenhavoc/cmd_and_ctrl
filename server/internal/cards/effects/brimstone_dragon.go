package effects

// Brimstone Dragon — Creature — Dragon {6}{R}{R}, 6/6 (Portal Second Age,
// #5625):
//
//	"Flying, haste"
//
// Flying and haste are engine keywords carried as printed data: it can be
// blocked only by creatures with flying or reach (CR 702.9b), and it can
// attack the turn it arrives (CR 702.10b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b20302d9-eb45-4895-9597-cf3af1abce56",
		Name:            "Brimstone Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
	})
}
