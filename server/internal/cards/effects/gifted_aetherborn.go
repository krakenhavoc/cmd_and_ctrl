package effects

// Gifted Aetherborn — Creature — Aetherborn Vampire {B}{B}, 2/3 (Aether Revolt, #5536):
// "Deathtouch, lifelink". Two printed keywords and nothing else.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0c6a6587-2b3f-40d1-a5f2-06128da0d52e",
		Name:            "Gifted Aetherborn",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch", "lifelink"},
	})
}
