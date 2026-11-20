package effects

// Moss Viper — Creature — Snake {G}, 1/1 (Theros Beyond Death, #5514):
//
//	"Deathtouch"
//
// Deathtouch is the printed keyword (Typhoid Rats' shape).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7d6f6ee4-77c2-4986-91c7-7dee962fc1a7",
		Name:            "Moss Viper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
	})
}
