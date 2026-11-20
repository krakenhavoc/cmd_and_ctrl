package effects

// Maritime Guard — Creature — Merfolk Soldier {1}{U}, 1/3 (Magic 2011,
// #5579):
//
// A vanilla creature: no rules text, so the spec is its identity.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0bb1c704-47f1-4d0c-8158-cfcd0b5e4d06",
		Name:         "Maritime Guard",
		Completeness: CompletenessFull,
	})
}
