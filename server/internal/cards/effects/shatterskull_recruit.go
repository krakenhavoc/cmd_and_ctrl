package effects

// Shatterskull Recruit — Creature — Giant Warrior Ally {3}{R}{R}, 4/4
// (Battle for Zendikar, #5547):
//
//	"Menace (This creature can't be blocked except by two or more
//	 creatures.)"
//
// Menace is an engine keyword carried as printed data: it can't be
// blocked except by two or more creatures (CR 702.111b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e2bc1600-eeec-443b-99e4-dd61497edd5b",
		Name:            "Shatterskull Recruit",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
	})
}
