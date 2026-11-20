package effects

// Merfolk of the Depths — Creature — Merfolk Soldier {4}{G/U}{G/U}, 4/2 (Gatecrash, #5565):
//
//	"Flash (You may cast this spell any time you could cast an instant.)"
//
// Flash is the engine keyword (CR 702.8).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "263df914-aea1-48dc-b517-7d515307789e",
		Name:            "Merfolk of the Depths",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
	})
}
