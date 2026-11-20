package effects

// Millennial Gargoyle — Artifact Creature — Gargoyle {4}, 2/2 (Gatecrash, #5565):
//
//	"Flying"
//
// Flying is the engine keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fe07a23a-e526-4277-8c39-6bae5355f0d4",
		Name:            "Millennial Gargoyle",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
