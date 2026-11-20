package effects

// Pyromantic Pilgrim — Creature — Human Wizard {2}{R}, 3/1 (Dominaria, #5530):
//
//	"Haste (This creature can attack and {T} as soon as it comes under your control.)"
//
// One printed keyword: it can attack the turn it comes under your control (CR 702.10b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "24d52cf1-d4f1-4b4c-a50b-87d3ba0aeebb",
		Name:            "Pyromantic Pilgrim",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
	})
}
