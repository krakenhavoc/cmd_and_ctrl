package effects

// Augmenting Automaton — Artifact Creature — Construct {1}, 1/1 (Aether Revolt,
// #5536):
//
//	"{1}{B}: This creature gets +1/+1 until end of turn."
//
// Tattered Apparition's self-pump row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5ec0e58c-99b5-4ebf-8053-f59c30b86ef7",
		Name:         "Augmenting Automaton",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{selfPumpRow("Augmenting Automaton", "{1}{B}", 1, 1)},
	})
}
