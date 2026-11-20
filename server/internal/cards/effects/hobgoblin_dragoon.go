package effects

// Hobgoblin Dragoon — Creature — Goblin Knight {2}{R/W}, 1/2 (Eventide, #5590):
//
//	"Flying, first strike"
//
// The two printed keywords (CR 702.9, CR 702.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3bbfceba-9f3a-4988-a009-13c9472138b1",
		Name:            "Hobgoblin Dragoon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
	})
}
