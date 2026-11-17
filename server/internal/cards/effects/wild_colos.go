package effects

// Wild Colos — Creature — Goat Beast {2}{R}, 2/2 (Urza's Destiny, #5621):
//
//	"Haste"
//
// One printed keyword, which the attack check reads: it can attack the
// turn it arrives (CR 702.10b).
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
