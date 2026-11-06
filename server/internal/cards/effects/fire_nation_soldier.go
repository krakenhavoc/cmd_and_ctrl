package effects

// Fire Nation Soldier — Creature — Human Soldier {2}{R}, 3/2 (Avatar: The Last Airbender
// Eternal, #5075):
//
//	"Haste (This creature can attack and {T} as soon as it comes under
//	 your control.)"
//
// The keyword rides PrintedKeywords; the combat engine does the rest.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6a15fad7-8171-4bce-bb39-fe3a8175f8fa",
		Name:            "Fire Nation Soldier",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
	})
}
