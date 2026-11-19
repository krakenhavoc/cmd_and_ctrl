package effects

// Mesa Unicorn — Creature — Unicorn {1}{W}, 2/2:
//
//	"Lifelink (Damage dealt by this creature also causes you to gain that
//	 much life.)"
//
// Lifelink is the engine's keyword (CR 702.15).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5101ff9e-4680-4d38-80a5-b4c21bc67356",
		Name:            "Mesa Unicorn",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
	})
}
