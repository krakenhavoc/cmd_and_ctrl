package effects

// Aardvark Sloth — Creature — Sloth Beast {3}{W}, 3/3 (Avatar: The Last Airbender
// Eternal, #5075):
//
//	"Lifelink (Damage dealt by this creature also causes you to gain
//	 that much life.)"
//
// The keyword rides PrintedKeywords; the combat engine does the rest.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1c385880-3eea-45c8-8e89-330ad8ad92b1",
		Name:            "Aardvark Sloth",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
	})
}
