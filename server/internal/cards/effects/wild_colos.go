package effects

// Wild Colos — Creature — Goat Beast {2}{R}, 2/2 (Urza's Destiny, #5621):
//
//	"Haste"
//
// Raging Goblin's haste on a 2/2 (CR 702.10a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cb6b8ce3-9f9d-418c-94b0-c4469a254938",
		Name:            "Wild Colos",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
	})
}
