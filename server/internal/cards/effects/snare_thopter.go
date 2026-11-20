package effects

// Snare Thopter — Artifact Creature — Thopter {4}, 3/2 (Kaladesh, #5539):
//
//	"Flying, haste"
//
// Printed keywords only; the engine reads each one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a3be0e16-6063-4cd0-87cb-606591a958c2",
		Name:            "Snare Thopter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
	})
}
