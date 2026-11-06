package effects

// Zhao, the Seething Flame — Legendary Creature — Human Soldier {4}{R}, 5/5 (Avatar: The Last Airbender
// Eternal, #5075):
//
//	"Menace (This creature can't be blocked except by two or more
//	 creatures.)"
//
// The keyword rides PrintedKeywords; the combat engine does the rest.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "92ce3e30-dde3-43bb-a3d6-ccd91ff18bc3",
		Name:            "Zhao, the Seething Flame",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
	})
}
