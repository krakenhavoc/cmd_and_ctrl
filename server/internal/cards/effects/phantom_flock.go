package effects

// Phantom Flock — Creature — Bird Soldier Spirit {3}{W}{W}, 0/0:
//
//	"Flying
//	 This creature enters with three +1/+1 counters on it.
//	 If damage would be dealt to this creature, prevent that damage. Remove a +1/+1 counter from this creature."
//
// ADR 0108 §8 (#1906): one of the Phantoms (phantoms.go). Damage dealt to
// it by any number of sources at once is all prevented and costs one
// counter; damage that can't be prevented is dealt and still costs one
// (CR 615.12); with no counter left the damage is still prevented (the
// rulings).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "bcded242-6e54-416f-bdc8-093211c50e3f",
		Name:            "Phantom Flock",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements:    phantomReplacements("Phantom Flock", 3),
	})
}
