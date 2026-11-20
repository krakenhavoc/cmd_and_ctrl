package effects

// Hand of Silumgar — Creature — Human Warrior {1}{B}, 2/1 (Dragons of Tarkir,
// #5551):
//
//	"Deathtouch (Any amount of damage this deals to a creature is enough to destroy it.)"
//
// Deathtouch is an engine keyword carried as printed data (CR 702.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e2714e1f-c883-41e0-bc2f-60e15462ef98",
		Name:            "Hand of Silumgar",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
	})
}
