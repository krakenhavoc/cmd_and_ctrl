package effects

// Griffin Sentinel — Creature — Griffin {2}{W}, 1/3 (Magic 2010, #5585):
//
//	"Flying
//	 Vigilance (Attacking doesn't cause this creature to tap.)"
//
// Serra Angel's keywords: flying (CR 702.9b) and vigilance (CR 702.20b), both carried as printed data.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "72df0b70-5c9f-44f3-96d5-81c95abc0065",
		Name:            "Griffin Sentinel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
	})
}
