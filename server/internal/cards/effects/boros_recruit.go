package effects

// Boros Recruit — Creature — Goblin Soldier {R/W}, 1/1 (Ravnica: City of
// Guilds, #5600):
//
//	"({R/W} can be paid with either {R} or {W}.)
//	 First strike"
//
// The hybrid pip is printed cost data the payment engine reads (CR 107.4e);
// first strike is the printed keyword (CR 702.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "00abb43b-550f-42d2-9049-61ac56ff5d8d",
		Name:            "Boros Recruit",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
	})
}
