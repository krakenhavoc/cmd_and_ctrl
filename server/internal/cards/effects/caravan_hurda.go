package effects

// Caravan Hurda — Creature — Giant {4}{W}, 1/5 (Zendikar, #5583):
//
//	"Lifelink (Damage dealt by this creature also causes you to gain
//	 that much life.)"
//
// Lifelink is the printed keyword (CR 702.15).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2c0b8cc3-b1c1-498d-b34c-f317c2c9468f",
		Name:            "Caravan Hurda",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
	})
}
