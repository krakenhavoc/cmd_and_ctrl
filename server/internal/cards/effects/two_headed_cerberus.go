package effects

// Two-Headed Cerberus — Creature — Dog {1}{R}{R}, 1/2 (Theros, #5561):
//
//	"Double strike (This creature deals both first-strike and regular combat damage.)"
//
// Printed keywords only; the engine reads them off the spec.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7b2bbd87-f288-44bd-a15c-1432a3e90cb3",
		Name:            "Two-Headed Cerberus",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
	})
}
