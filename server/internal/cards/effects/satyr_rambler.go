package effects

// Satyr Rambler — Creature — Satyr {1}{R}, 2/1 (Theros, #5561):
//
//	"Trample"
//
// Printed keywords only; the engine reads them off the spec.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6ea0d274-564b-4785-801d-b3272a7b6bfa",
		Name:            "Satyr Rambler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
