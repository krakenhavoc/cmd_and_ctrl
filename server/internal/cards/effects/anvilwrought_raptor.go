package effects

// Anvilwrought Raptor — Artifact Creature — Bird {4}, 2/1 (Theros, #5561):
//
//	"Flying
//	 First strike (This creature deals combat damage before creatures without first strike.)"
//
// Printed keywords only; the engine reads them off the spec.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d87ebd8c-ad2c-45c9-8c97-f046aeda861c",
		Name:            "Anvilwrought Raptor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
	})
}
