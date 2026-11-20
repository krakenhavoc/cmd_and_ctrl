package effects

// Hitchclaw Recluse — Creature — Spider {2}{G}, 1/4 (Magic Origins, #5549):
//
//	"Reach"
//
// Giant Spider's shape: reach feeds Game.CanBlockLocked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "241ec1fc-cc50-468e-8ee3-f95a82a3e07b",
		Name:            "Hitchclaw Recluse",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
	})
}
