package effects

// Watercourser — Creature — Elemental {2}{U}, 2/3 (Magic 2013, #5568):
//
//	"{U}: This creature gets +1/-1 until end of turn."
//
// Thornling's +1/-1 row for {U}: any number of times, the turn it
// arrives included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "65dd4ae3-432c-4d57-ac14-4d827262ea0b",
		Name:         "Watercourser",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{thisGetsRow("{U}", 1, -1, "Watercourser")},
	})
}
