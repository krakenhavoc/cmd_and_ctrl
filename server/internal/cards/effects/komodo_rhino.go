package effects

// Komodo Rhino — Creature — Lizard Rhino {3}{R}, 5/2 (Avatar: The Last Airbender
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
		OracleID:        "7fe5cad0-8db8-4cb3-9d61-bda84477654d",
		Name:            "Komodo Rhino",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
