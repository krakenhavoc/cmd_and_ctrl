package effects

// Titanium Golem — Artifact Creature — Golem {5}, 3/3 (Mirrodin, #5606):
//
//	"{1}{W}: This creature gains first strike until end of turn."
//
// Malachite Golem's row with first strike: a layer 6 grant pinned to
// the source until cleanup (CR 514.2). Granted before combat damage,
// it puts the Golem in the first-strike damage step (CR 510.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "84fa0c22-4e11-4d57-9a45-633f0e1686de",
		Name:         "Titanium Golem",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{udsB3ThisGainsRow("{1}{W}", "first strike", "Titanium Golem")},
	})
}
