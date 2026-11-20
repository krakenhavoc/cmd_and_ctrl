package effects

// Vampire Revenant — Creature — Vampire Spirit {3}{B}, 3/1 (Rivals of Ixalan, #5531):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data: only a creature
// with flying or reach can block it (CR 702.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2f7305c1-51d2-4698-89d5-c64a57ecff80",
		Name:            "Vampire Revenant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
