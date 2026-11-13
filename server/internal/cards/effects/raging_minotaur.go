package effects

// Raging Minotaur — Creature — Minotaur Berserker {2}{R}{R}, 3/3 (Portal, #5630):
//
//	"Haste"
//
// Haste is an engine keyword carried as printed data: it can attack and
// tap the turn it arrives (CR 702.10b, CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4cfd04b3-57bd-4860-8b47-57a0fd0a23fa",
		Name:            "Raging Minotaur",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
	})
}
