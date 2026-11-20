package effects

// Zealous Guardian — Creature — Kithkin Soldier {W/U}, 1/1 (Shadowmoor,
// #5591):
//
//	"Flash"
//
// Flash is a printed keyword (CR 702.8a): it is cast any time it could
// cast an instant.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "508fdfb1-87d6-4342-a05a-01158f253896",
		Name:            "Zealous Guardian",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
	})
}
