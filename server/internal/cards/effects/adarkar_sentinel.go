package effects

// Adarkar Sentinel — Artifact Creature — Soldier {5}, 3/3 (Ice Age,
// #5636):
//
//	"{1}: This creature gets +0/+1 until end of turn."
//
// The shared selfPumpRow at +0/+1 for {1}: no tap, so any number of
// activations, summoning sickness or not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bd372a53-48f9-4e9a-ab00-0f9c4606d452",
		Name:         "Adarkar Sentinel",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{selfPumpRow("Adarkar Sentinel", "{1}", 0, 1)},
	})
}
