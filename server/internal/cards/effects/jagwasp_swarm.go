package effects

// Jagwasp Swarm — Creature — Insect {3}{B}, 3/2 (Worldwake, #5582):
//
//	"Flying"
//
// The printed keywords are engine keywords carried as data, as on
// Cloud Djinn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "376d90f6-2ceb-4d9e-a04d-3dd5bafd1e8b",
		Name:            "Jagwasp Swarm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
