package effects

// Squirrelanoids — {B} Creature — Squirrel Mutant 1/1 (Teenage Mutant Ninja
// Turtles, #4457):
//
//	"Deathtouch"
//
// Typhoid Rats's printed keyword and nothing else.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "315ddc7a-c39c-43d9-b134-23acbd865e3e",
		Name:            "Squirrelanoids",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
	})
}
