package effects

// Elite Vanguard — Creature — Human Soldier {W}, 2/1 (Magic 2010, #5585):
//
// A vanilla creature: no rules text, so the spec is its identity.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b3b9a87d-cb95-435c-90b6-037406cab32e",
		Name:         "Elite Vanguard",
		Completeness: CompletenessFull,
	})
}
