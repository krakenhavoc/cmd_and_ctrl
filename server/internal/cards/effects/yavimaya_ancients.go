package effects

// Yavimaya Ancients — Creature — Treefolk {3}{G}{G}, 2/7 (Alliances,
// #5634):
//
//	"{G}: This creature gets +1/-2 until end of turn."
//
// A layer 7c change (CR 613.4c) that ends at cleanup (CR 514.2). Each
// activation trades two toughness for a point of power, so a fourth in
// one turn leaves it at 0 toughness and it is put into its owner's
// graveyard (CR 704.5f). Built by thisGetsRow.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0e91a873-a680-4ca2-b073-26aa39376d77",
		Name:         "Yavimaya Ancients",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{thisGetsRow("{G}", 1, -2, "Yavimaya Ancients")},
	})
}
