package effects

// Raging Goblin — Creature — Goblin Berserker {R}, 1/1 (Portal, #5630):
//
//	"Haste"
//
// Haste is an engine keyword carried as printed data: it can attack and
// tap the turn it arrives (CR 702.10b, CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "30997b43-fc13-41d3-8064-1ccc2cb6fd2b",
		Name:            "Raging Goblin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
	})
}
