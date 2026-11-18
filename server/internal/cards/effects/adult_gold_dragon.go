package effects

// Adult Gold Dragon — Creature — Dragon {3}{R}{W}, 4/3 (Adventures in the Forgotten Realms):
//
//	Flying, lifelink, haste
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a361fdd1-7f97-4aed-8e24-ab2a03387f73",
		Name:            "Adult Gold Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink", "haste"},
	})
}
