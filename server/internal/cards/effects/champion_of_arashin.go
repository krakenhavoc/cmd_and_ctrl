package effects

// Champion of Arashin — Creature — Dog Warrior {3}{W}, 3/2 (Dragons of Tarkir,
// #5551):
//
//	"Lifelink (Damage dealt by this creature also causes you to gain that much life.)"
//
// Lifelink is an engine keyword carried as printed data (CR 702.15).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f688863f-b8eb-4143-8ea5-0b234cc91bd0",
		Name:            "Champion of Arashin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
	})
}
