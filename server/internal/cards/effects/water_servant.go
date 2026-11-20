package effects

// Water Servant — Creature — Elemental {2}{U}{U}, 3/4 (Magic 2011, #5579):
//
//	"{U}: This creature gets +1/-1 until end of turn.
//	 {U}: This creature gets -1/+1 until end of turn."
//
// Thornling's two trades through the shared thisGetsRow, each a layer 7c
// change that ends at cleanup (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f8c32c80-9693-4ae4-a387-21ec8a3cdd40",
		Name:         "Water Servant",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			thisGetsRow("{U}", 1, -1, "Water Servant"),
			thisGetsRow("{U}", -1, 1, "Water Servant"),
		},
	})
}
