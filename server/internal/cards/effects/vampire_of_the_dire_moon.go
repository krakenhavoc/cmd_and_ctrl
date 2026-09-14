package effects

// Vampire of the Dire Moon — Creature — Vampire {B}, 1/1 (EDHREC
// rank 3319):
//
//	"Deathtouch
//	 Lifelink"
//
// One black mana for a deathtouch lifelinker. Both keywords ride
// PrintedKeywords; the engine's combat path does the rest.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5a938371-8428-48e3-85ed-6a84a39b29c6",
		Name:            "Vampire of the Dire Moon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch", "lifelink"},
	})
}
