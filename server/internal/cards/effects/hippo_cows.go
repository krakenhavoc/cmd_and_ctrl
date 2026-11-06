package effects

// Hippo-Cows — Creature — Hippo Ox {4}{G}, 5/4 (Avatar: The Last Airbender
// Eternal, #5075):
//
//	"Trample (This creature can deal excess combat damage to the player
//	 it's attacking.)"
//
// The keyword rides PrintedKeywords; the combat engine does the rest.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f66fca2b-6c44-4079-b29e-4d18ef030a0c",
		Name:            "Hippo-Cows",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
