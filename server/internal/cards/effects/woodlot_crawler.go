package effects

// Woodlot Crawler — Creature — Insect {U}{B}, 2/1:
//
//	"Forestwalk (This creature can't be blocked as long as defending
//	 player controls a Forest.)
//	 Protection from green"
//
// Forestwalk is landwalk, read at the block declaration (CR 702.14c);
// protection from green is the engine's DEBT keyword (CR 702.16), as on
// Zombie Outlander.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e5b11e43-fc7c-4cea-a44a-8af3f124e29e",
		Name:            "Woodlot Crawler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"forestwalk", "protection from green"},
	})
}
