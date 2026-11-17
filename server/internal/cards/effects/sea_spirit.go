package effects

// Sea Spirit — Creature — Elemental Spirit {4}{U}, 2/3 (Ice Age,
// #5636):
//
//	"{U}: This creature gets +1/+0 until end of turn."
//
// The shared selfPumpRow at +1/+0 for {U}: no tap, so any number of
// activations, summoning sickness or not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0ebd8ab3-d4b7-480e-b7b1-267078edddf2",
		Name:         "Sea Spirit",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{selfPumpRow("Sea Spirit", "{U}", 1, 0)},
	})
}
