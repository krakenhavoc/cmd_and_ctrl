package effects

// Skyswirl Harrier — Creature — Bird {4}{W}, 3/4 (Kaladesh, #5539):
//
//	"Flying"
//
// Printed keywords only; the engine reads each one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0feeeb19-7c90-4009-9a71-226a443b5c3f",
		Name:            "Skyswirl Harrier",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
