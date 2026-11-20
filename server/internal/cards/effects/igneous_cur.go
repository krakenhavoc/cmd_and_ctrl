package effects

// Igneous Cur — Creature — Elemental Dog {1}{R}, 1/2 (Core Set 2021, #5511):
//
//	"{1}{R}: This creature gets +2/+0 until end of turn."
//
// Hematite Golem's pump (thisGetsRow): a layer 7c boost on the source
// that ends at cleanup, any number of times.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0d899095-cef7-43d2-aaf7-eb73b62f5ec8",
		Name:         "Igneous Cur",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{thisGetsRow("{1}{R}", 2, 0, "Igneous Cur")},
	})
}
