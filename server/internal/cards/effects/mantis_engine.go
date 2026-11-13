package effects

// Mantis Engine — Artifact Creature — Insect {5}, 3/3 (Urza's Destiny,
// #5621):
//
//	"{2}: This creature gains flying until end of turn.
//	 {2}: This creature gains first strike until end of turn."
//
// Two rows of the shared udsB2ThisGainsRow: each is a layer-6 grant on
// the Engine that ends in the cleanup step (CR 514.2). Neither cost has
// {T}, so both can be activated the turn it arrives.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "57bea272-5b38-444e-acc2-6104af2ece22",
		Name:         "Mantis Engine",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			udsB2ThisGainsRow("{2}", "flying", "Mantis Engine"),
			udsB2ThisGainsRow("{2}", "first strike", "Mantis Engine"),
		},
	})
}
