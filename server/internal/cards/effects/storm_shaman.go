package effects

// Storm Shaman — Creature — Human Cleric Shaman {2}{R}, 0/4 (Alliances,
// #5634):
//
//	"{R}: This creature gets +1/+0 until end of turn."
//
// A layer 7c boost (CR 613.4c) that ends at cleanup (CR 514.2). With no
// {T} in the cost it can be activated any number of times, the turn it
// arrives included. Built by thisGetsRow.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "65e07c0f-4eff-4b38-9785-63684d67c0d8",
		Name:         "Storm Shaman",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{thisGetsRow("{R}", 1, 0, "Storm Shaman")},
	})
}
