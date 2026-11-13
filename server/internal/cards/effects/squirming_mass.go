package effects

// Squirming Mass — Creature — Horror {1}{B}, 1/1 (Urza's Destiny, #5621):
//
//	"Fear (This creature can't be blocked except by artifact creatures
//	 and/or black creatures.)"
//
// Razortooth Rats' fear on a 1/1: only artifact creatures and black
// creatures can block it (CR 702.36b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2c526c61-6cdb-440c-b4c8-b6d1503cef73",
		Name:            "Squirming Mass",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"fear"},
	})
}
