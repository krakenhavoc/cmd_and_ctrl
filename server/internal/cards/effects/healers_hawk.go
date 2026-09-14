package effects

// Healer's Hawk — Creature — Bird {W}, 1/1 (EDHREC rank 3839):
//
//	"Flying
//	 Lifelink (Damage dealt by this creature also causes you to gain
//	 that much life.)"
//
// A one-drop evasive lifelinker. Both keywords ride PrintedKeywords;
// the combat engine does the rest.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "28a52ba1-95da-44e1-8ac5-0dc23c902394",
		Name:            "Healer's Hawk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
	})
}
