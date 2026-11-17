package effects

// Folk of the Pines — Creature — Dryad {4}{G}, 2/5 (Ice Age, #5636):
//
//	"{1}{G}: This creature gets +1/+0 until end of turn."
//
// The shared selfPumpRow at +1/+0 for {1}{G}: no tap, so any number of
// activations, summoning sickness or not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6cd8cf2a-972f-41b7-ab0c-b4953accc2fe",
		Name:         "Folk of the Pines",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{selfPumpRow("Folk of the Pines", "{1}{G}", 1, 0)},
	})
}
