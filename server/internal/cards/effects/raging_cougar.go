package effects

// Raging Cougar — Creature — Cat {2}{R}, 2/2 (Portal, #5630):
//
//	"Haste"
//
// Haste is an engine keyword carried as printed data: it can attack and
// tap the turn it arrives (CR 702.10b, CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "695ebb35-55fd-4ed1-8cbe-e3fc1223115b",
		Name:            "Raging Cougar",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
	})
}
