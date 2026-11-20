package effects

// Alaborn Grenadier — Creature — Human Soldier {W}{W}, 2/2 (Portal Second Age,
// #5625):
//
//	"Vigilance"
//
// Vigilance is an engine keyword carried as printed data: attacking
// doesn't tap it (CR 702.20b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "69c4429e-a8a4-49c3-bb43-9dce4cc6b90a",
		Name:            "Alaborn Grenadier",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
}
