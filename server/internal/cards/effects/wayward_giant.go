package effects

// Wayward Giant — Creature — Giant {4}{R}, 4/5 (Kaladesh, #5539):
//
//	"Menace"
//
// Printed keywords only; the engine reads each one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "124cf15a-587a-4f1f-a943-f462072f2098",
		Name:            "Wayward Giant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
	})
}
