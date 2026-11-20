package effects

// Malachite Golem — Artifact Creature — Golem {6}, 5/3 (Mirrodin, #5606):
//
//	"{1}{G}: This creature gains trample until end of turn."
//
// A layer 6 grant pinned to the source until cleanup (CR 514.2),
// the shape Endling's keyword rows use.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "54e71732-6575-48f4-838b-35d44eba51cf",
		Name:         "Malachite Golem",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{udsB3ThisGainsRow("{1}{G}", "trample", "Malachite Golem")},
	})
}
