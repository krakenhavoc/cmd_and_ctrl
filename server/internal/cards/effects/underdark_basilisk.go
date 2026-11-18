package effects

// Underdark Basilisk — Creature — Basilisk {1}{G}, 1/2 (Adventures in the Forgotten Realms):
//
//	Deathtouch
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3b61b6da-44ed-4766-8134-fca37afd3f2a",
		Name:            "Underdark Basilisk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
	})
}
