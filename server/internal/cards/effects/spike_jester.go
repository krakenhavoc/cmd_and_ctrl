package effects

// Spike Jester — Creature — Goblin Warrior {B}{R}, 3/1:
//
//	"Haste"
//
// Haste lets it attack the turn it arrives (CR 702.10b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "29f22168-21aa-4ca9-ab72-b09396a12c76",
		Name:            "Spike Jester",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
	})
}
