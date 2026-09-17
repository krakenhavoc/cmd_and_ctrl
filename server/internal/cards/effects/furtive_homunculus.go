package effects

// Furtive Homunculus — "Skulk (This creature can't be blocked by creatures
// with greater power.)"
func init() {
	Register(Spec{
		OracleID:        "6b59ffa2-5c8b-4549-aa09-2f2a5cabc5af",
		Name:            "Furtive Homunculus",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"skulk"},
	})
}
