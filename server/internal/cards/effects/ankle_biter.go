package effects

// Ankle Biter — Creature — Snake {G}, 1/1 (Outlaws of Thunder Junction,
// #4468):
//
//	"Deathtouch"
//
// Deathtouch is the engine's combat keyword (Typhoid Rats' row).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ee8f16ba-cbf0-4286-b26c-d78dd35bee17",
		Name:            "Ankle Biter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
	})
}
