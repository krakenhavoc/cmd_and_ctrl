package effects

// Havoc Devils — Creature — Devil {2}{R}{R}, 4/3 (Core Set 2019, #5527):
//
//	"Trample (This creature can deal excess combat damage to the player or planeswalker it's attacking.)"
//
// The whole text is printed keywords; the engine reads them off
// PrintedKeywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "be1c2607-091a-41e7-ae4c-73c7e5689739",
		Name:            "Havoc Devils",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
